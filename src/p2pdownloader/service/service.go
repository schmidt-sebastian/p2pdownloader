package service

import (
	"bytes"
	"p2pdownloader/protocol"
	"errors"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"v.io/v23"
	"v.io/v23/context"
	"v.io/v23/naming"
	"time"
	"v.io/v23/rpc"
	"sync"
	"fmt"
)

type Logic struct {
	ctx                 *context.T
	localEndpoint       string
	mountableRoot       string
	readyToDispatchLock *sync.Mutex
	readyToDispatch     []Chunk
	readyToDownload     chan Chunk
	readyToTransfer     chan Chunk
}

type Chunk struct {
	owner      string
	url        string
	start      int
	length     int
	downloaded int
	data       *bytes.Buffer
}

const (
	chunkSize = 0x1 << 20 // 1 megabyte
	numberThreads = 3
)

var (
	emptyKilobyte = make([]byte, 1024)
)

func Make(ctx *context.T, localEndpoint, mountableRoot string) *Logic {
	logic := &Logic{
		readyToDispatchLock: &sync.Mutex{},
		readyToDispatch: make([]Chunk, 0),
		readyToDownload : make(chan Chunk),
		readyToTransfer  : make(chan Chunk),
		localEndpoint: localEndpoint,
		mountableRoot: mountableRoot,
		ctx:           ctx,
	}

	return logic
}

// Starts the goroutines that manage all background work
func (s *Logic) Start() {
	for i := 0; i < numberThreads; i++ {
		go s.downloadChunks()
	}

	go s.gatherDownloads()
	go s.transferRemoteChunks()
}

// Shuts down all background tasks by closing the channels
func (s *Logic) Shutdown() {
	close(s.readyToDownload)
	close(s.readyToTransfer)
}

// Registers a new download in the local queue
func (s *Logic) AddDownload(url string) {
	s.ctx.Infof("Starting download %v", url)

	fileSize, err := retrieveLength(url)

	if err != nil {
		s.ctx.Errorf("Could not determine download size")
		return
	}

	s.prepareFile(url, fileSize)
	s.registerDownload(url, fileSize)
}

// RPC handler that retrieves a pending download from the local queue
func (s *Logic) GetChunk(_ *context.T, _ rpc.ServerCall) (activeDownloads downloader.Chunk, _ error) {
	s.readyToDispatchLock.Lock()
	defer s.readyToDispatchLock.Unlock()

	if len(s.readyToDispatch) > 0 {
		s.ctx.Info("Dispatching chunk to peer")
		result := convertToRemoteChunk(s.readyToDispatch[0])
		s.readyToDispatch = s.readyToDispatch[1:]
		return result, nil
	} else {
		return downloader.Chunk{}, errors.New("No chunk available")
	}
}

// RPC handler that sends the byte data back to the originating peer
func (s *Logic) TransferChunk(_ *context.T, _ rpc.ServerCall, downloadedChunk downloader.Chunk) error {
	s.writeChunk(convertToLocalChunk(downloadedChunk))
	return nil
}

// Converts the RPC representation of 'chunk' to the local object
func convertToLocalChunk(input downloader.Chunk) Chunk {
	chunk := Chunk{
		url:    input.Url,
		owner:  input.Owner,
		start:  int(input.Offset),
		length: int(input.Length),
		data:   new(bytes.Buffer),
	}
	chunk.data.Write(input.Data)
	return chunk
}

// Converts the local chunk representation to the RPC representation
func convertToRemoteChunk(input Chunk) downloader.Chunk {
	chunk := downloader.Chunk{
		Url:    input.url,
		Offset: int32(input.start),
		Length: int32(input.length),
		Data:   input.data.Bytes(),
		Owner:  input.owner,
	}
	return chunk
}

// Issues a HTTP HEAD request and returns the download size
func retrieveLength(url string) (int, error) {
	header, err := http.Head(url)

	if err != nil {
		return 0, err
	}

	lengthHeader := header.Header.Get("Content-Length")

	if len(lengthHeader) > 0 {
		return strconv.Atoi(lengthHeader)
	} else {
		return 0, errors.New("Cannot compute length")
	}
}

// GoRoutine that performs the downloads of the queued up chunks on this peer
func (s *Logic) downloadChunks() {
	for {
		chunk, ok := <-s.readyToDownload

		if ok {
			s.ctx.Infof("Downloading chunk %v for %v", chunk.start / chunkSize, chunk.url)
			request, err := http.NewRequest("GET", chunk.url, nil)
			client := &http.Client{}

			if err != nil {
				s.ctx.Errorf("Error encountered: %v", err.Error())
				continue
			}

			request.Header.Add("Range", fmt.Sprintf("bytes=%v-%v", chunk.start, chunk.start + chunk.length - 1))
			resp, err := client.Do(request)
			n, err := chunk.data.ReadFrom(resp.Body)
			defer resp.Body.Close()

			if err != nil {
				s.ctx.Errorf("Error encountered while reading: %v", err.Error())
				continue
			}

			s.ctx.Infof("Downloaded %v bytes for chunk %v with status %v for %v", n,
				chunk.start / chunkSize, resp.Status, chunk.url)
			s.readyToTransfer <- chunk
		} else {
			s.ctx.Info("Shutting down")
			return
		}
	}

}

// Persist the byte data in 'chunk' to disk
func (s *Logic) writeChunk(chunk Chunk) {
	url, _ := url.Parse(chunk.url)
	filename := url.Path[strings.LastIndex(url.Path, "/") + 1:]

	out, err := os.OpenFile(filename, os.O_RDWR, 0666)
	if err != nil {
		s.ctx.Errorf("Could not open output file")
		return
	}
	defer out.Close()

	_, err = out.Seek(int64(chunk.start), 0)

	if err != nil {
		s.ctx.Errorf("Could not seek in output file")
		return
	}

	_, err = out.Write(chunk.data.Bytes())

	if err != nil {
		s.ctx.Errorf("Could not write in output file")
		return
	}

	s.ctx.Infof("Wrote chunk ")
}

// Send downloaded chunks back to their originating peer
func (s *Logic) transferRemoteChunks() {
	for {
		chunk, ok := <-s.readyToTransfer
		if ok {
			s.ctx.Infof("Transferring chunk to %v", chunk.owner)
			err := downloader.DownloaderClient(chunk.owner).TransferChunk(s.ctx, convertToRemoteChunk(chunk))
			if err != nil {
				s.ctx.Errorf("Could not transfer chunk %v", err.Error())
			}
		} else {
			s.ctx.Info("Shutting down")
			return
		}
	}
}

// Creates a blank file with the size of the final download
func (s *Logic) prepareFile(urlpath string, fileSize int) {
	url, err := url.Parse(urlpath)

	if err != nil {
		s.ctx.Errorf("Could not parse output file")
		return
	}

	filename := url.Path[strings.LastIndex(url.Path, "/") + 1:]

	out, err := os.Create(filename)
	defer out.Close()

	if err != nil {
		s.ctx.Errorf("Could not create output file")
		return
	}

	for written := 0; written < fileSize; written++ {
		length := len(emptyKilobyte)

		if length + written > fileSize {
			length = fileSize - written
		}

		n, err := out.Write(emptyKilobyte[:length])

		if err != nil {
			s.ctx.Errorf("Could not write to empty file")
			return
		}

		written += n
	}

	s.ctx.Info("Prepared file")
}

// Creates the pending chunks for the download specified via 'url'
func (s *Logic) registerDownload(url string, fileSize int) {
	numberChunks := fileSize / chunkSize + int(math.Min(float64(fileSize % chunkSize), 1.0))

	s.readyToDispatchLock.Lock()
	defer s.readyToDispatchLock.Unlock()

	for i := 0; i < numberChunks; i++ {
		start := i * chunkSize
		length := int(math.Min(float64(fileSize), float64(i + 1) * chunkSize)) - start
		s.readyToDispatch = append(s.readyToDispatch, Chunk{url: url, start: start, length: length,
			data: new(bytes.Buffer), owner: s.localEndpoint})
	}

	s.ctx.Infof("Added %v chunks", numberChunks)
}

// Discovers all peers that can be used for distributed downloads
func (s *Logic) findAllPeers() ([]string, error) {
	ns := v23.GetNamespace(s.ctx)

	globReply, err := ns.Glob(s.ctx, s.mountableRoot + "/*")
	if err != nil {
		s.ctx.Errorf("Glob for namespace failed: %v", err)
		return nil, err
	}
	var servers []string
	for reply := range globReply {
		switch v := reply.(type) {
		case *naming.GlobReplyEntry:
			servers = append(servers, v.Value.Name)
		case *naming.GlobReplyError:
			s.ctx.Errorf("Glob %s can't be traversed: %s", v.Name, v.Value.Error)
		}
	}
	return servers, nil
}

// Asks other peers if they have downloads that the local peer can serve
func (s *Logic) gatherDownloads() {
	for true {
		if len(s.readyToDownload) < numberThreads {
			time.Sleep(time.Second)
			servers, err := s.findAllPeers()
			if err != nil {
				s.ctx.Errorf("Could not find servers %v", err.Error())
				continue
			}

			for _, serverName := range servers {
				s.ctx.Infof("Gathering downloads from %v", serverName)
				chunk, err := downloader.DownloaderClient(serverName).GetChunk(s.ctx)
				if err == nil {
					s.ctx.Infof("Adding external download to queue")
					s.readyToDownload <- convertToLocalChunk(chunk)
				}
			}
		}
	}
}
