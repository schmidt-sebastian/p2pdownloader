# P2PDownloader -- peer-to-peer assisted file downloading

A peer to peer file downloader that spreads downloads between multiple clients.

## Features

P2PDownloader can be started in two different modes: As a pure client that just assists other
downloads and as a peer that adds a new download, while at the same time also assisting other 
clients with their existing downloads.
All downloads are split into equal-size chunks and distributed among all peers. By spreading
the downloads, the WAN bandwidth of multiple peers can be combined, while all local transfers
between peers are done via the local network.

Downloader supports resumable HTTP downloads only (per HTTP range header).

## Installation

go install p2pdownloader/server

## Usage

To start a new peer:

./server --v23.credentials=/path/to/cred

To start a new peer and download a new file:

./server --v23.credentials=/path/to/cred -download http://...

'/path/to/cred' is the location of the credentials folder that you created with 'principal create'. 
Please not that you need to obtain a blessing from the global mountable before you can use these 
credentials (using 'principal seekblessings'). 

### How it works

Peers that get started with a download option determine the download size via 'HTTP Head' requests
and prepare a blank file as the destination of the download. They also prepare a list of 1 MB chunks
that are used to distribute the download load between all available peers.

Peers are uniquely registered in the global mountable using the name tmp/downloader/<UUID> and are
discovered via GLOB search (tmp/downloader/*). All peers download multiple chunks in parallel and
ask each other (and themselves) for chunks to download.

P2PDownloader uses two RPCs to distribute both the work and the results:

```
type Downloader interface {
  GetChunk() (newChunk Chunk | error)
  TransferChunk(downloadedChunk Chunk) error
}
```

'GetChunk' is used to transfer a queued-up and empty chunk from one peer to another. This only 
applies to chunks that are not yet downloading on any peer and allows them to distribute their
workload.

'TransferChunk' is used to transfer the downloaded data back to the originating peer who then 
persists it to disk.

### Caveats

1) There is currently no clear indication for finished downloads. All files are fully downloaded
   when the console outputs indicates no further progress.
   
2) Since P2PDownloader uses a global mountable, all downloads are spread between all active peers.
   This can include peers that are not on your local network. This can be circumvented by 
   specifying a unique mountable path using the "-mountable" flag.

