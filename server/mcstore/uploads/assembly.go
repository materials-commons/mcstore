package uploads

import "io"

// assembleRequest assembles all the chunks together into a single destination.
func assembleRequest(requestChunks chunkSupplier, destination io.Writer) error {
	chunks, err := requestChunks.chunks()
	if err != nil {
		return err
	}

	for _, chunk := range chunks {
		if err := writeChunk(chunk, destination); err != nil {
			return err
		}
	}
	return nil
}

// writeChunk writes a single chunk into the destination.
func writeChunk(chunk chunk, destination io.Writer) error {
	switch source, err := chunk.Reader(); {
	case err != nil:
		return err
	default:
		if closer, ok := source.(io.ReadCloser); ok {
			defer closer.Close()
		}
		_, err = io.Copy(destination, source)
		return err
	}
}
