# peer-drop

Peer A                       Signaling Server                   Peer B
  |----> Connect (Signaling) ----->|                              |
  |----> Create SDP Offer -------->|                              |
  |                                |<------ SDP Offer ------------|
  |                                |------ SDP Answer ----------->|
  |<----- Exchange ICE Candidates -->|<----- Exchange ICE Candidates -->|
  |                                |                              |
  |<------ Establish P2P Connection ---------------------------->|
  |<--------------- File Sharing over WebRTC Data Channel ------>|
