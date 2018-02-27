package vmx

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"io"
)

func crypto(w io.Writer, rReal io.Reader, p int64, blkSize int, key []byte) (readCount int, writeCount int, err error) {

	iv := Uint128{H: 0, L: uint64(p)}

	// округляем blkSize
	if blkSize%aes.BlockSize > 0 {
		blkSize = blkSize - (blkSize % aes.BlockSize)
	}
	r := bufio.NewReaderSize(rReal, blkSize)

	var block cipher.Block
	block, err = aes.NewCipher(key)
	if err != nil {
		return
	}
	emb := cipher.NewCBCEncrypter(block, iv.Bytes())
	buff := make([]byte, blkSize)
	padbuff := bytes.Repeat([]byte{byte(aes.BlockSize)}, aes.BlockSize)
	var readN, writeN, readNTmp int

	canRun := true
	for canRun {
		readN, err = io.ReadFull(r, buff)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				canRun = false
			} else {
				return
			}
		}
		readCount += readN
		if readN%aes.BlockSize == 0 {
			emb.CryptBlocks(buff[0:readN], buff[0:readN])
			writeN, err = w.Write(buff[0:readN])
			writeCount += writeN
			if err != nil {
				return
			}
		} else {
			readNTmp = readN - (readN % aes.BlockSize)
			padbuff = buff[readNTmp:readN]
			emb.CryptBlocks(buff[0:readNTmp], buff[0:readNTmp])
			writeN, err = w.Write(buff[0:readNTmp])
			writeCount += writeN
			if err != nil {
				return
			}
		}
	}

	padLen := aes.BlockSize - len(padbuff)
	padbuff = append(padbuff, bytes.Repeat([]byte{byte(padLen)}, padLen)...)
	emb.CryptBlocks(padbuff, padbuff)
	writeN, err = w.Write(padbuff)
	writeCount += writeN
	if err != nil {
		return
	}
	return readCount, writeCount, nil
}
