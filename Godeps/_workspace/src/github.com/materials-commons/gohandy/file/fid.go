package file

// FID represents a file identifier. On Linux it is the INode. On Windows
// it is the file index and volume serial number.
type FID struct {
	IDLow  uint64
	IDHigh uint64
}

// Equal compares to FIDs for equality. It can be used to check
// if a file has changed.
func Equal(fid1, fid2 FID) bool {
	return fid1.IDLow == fid2.IDLow && fid1.IDHigh == fid2.IDHigh
}
