package hedx

func (i *Index) Chunk(offset, size int64, timestamp float64) {
	i.fill(TypeChunk, offset, size, timestamp, false)
}

func (i *Index) ChunkKey(offset, size int64, timestamp float64) {
	i.fill(TypeKey, offset, size, timestamp, false)
}

func (i *Index) ChunkEOF(offset, size int64, timestamp float64) {
	i.fill(TypeEOF, offset, size, timestamp, false)
}

func (i *Index) IFrame(offset, size int64, timestamp float64) {
	i.fill(TypeIframe, offset, size, timestamp, true)
}

func (i *Index) IFrameKey(offset, size int64, timestamp float64) {
	i.fill(TypeKey, offset, size, timestamp, true)
}

func (i *Index) IFrameEOF(offset, size int64, timestamp float64) {
	i.fill(TypeEOF, offset, size, timestamp, true)
}

func (i *Index) fill(typ uint8, offset, size int64, timestamp float64, iframe bool) {
	i.Type = typ
	i.OffsetBytes = uint64(offset)
	i.SizeBytes = uint64(size)
	i.TimeStampUsec = uint64(1000000 * timestamp)
	i.Flags = 0x00
	if iframe {
		i.Flags = 0x01
	}
	i.Reserved = uint32(0)
}
