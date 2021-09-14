// Code generated by fastssz. DO NOT EDIT.
// Hash: 5f6be7a7e7933faa353dc4da7cc5ad7cefed308f93d54fb14e23023cbefb6a54
package carrier_rpc_v1

import (
	ssz "github.com/ferranbt/fastssz"
)

// MarshalSSZ ssz marshals the GossipTestData object
func (g *GossipTestData) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(g)
}

// MarshalSSZTo ssz marshals the GossipTestData object to a target array
func (g *GossipTestData) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(20)

	// Offset (0) 'GetData'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(g.Data)

	// Field (1) 'Count'
	dst = ssz.MarshalUint64(dst, g.Count)

	// Field (2) 'Step'
	dst = ssz.MarshalUint64(dst, g.Step)

	// Field (0) 'GetData'
	if len(g.Data) > 16777216 {
		err = ssz.ErrBytesLength
		return
	}
	dst = append(dst, g.Data...)

	return
}

// UnmarshalSSZ ssz unmarshals the GossipTestData object
func (g *GossipTestData) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 20 {
		return ssz.ErrSize
	}

	tail := buf
	var o0 uint64

	// Offset (0) 'GetData'
	if o0 = ssz.ReadOffset(buf[0:4]); o0 > size {
		return ssz.ErrOffset
	}

	if o0 < 20 {
		return ssz.ErrInvalidVariableOffset
	}

	// Field (1) 'Count'
	g.Count = ssz.UnmarshallUint64(buf[4:12])

	// Field (2) 'Step'
	g.Step = ssz.UnmarshallUint64(buf[12:20])

	// Field (0) 'GetData'
	{
		buf = tail[o0:]
		if len(buf) > 16777216 {
			return ssz.ErrBytesLength
		}
		if cap(g.Data) == 0 {
			g.Data = make([]byte, 0, len(buf))
		}
		g.Data = append(g.Data, buf...)
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the GossipTestData object
func (g *GossipTestData) SizeSSZ() (size int) {
	size = 20

	// Field (0) 'GetData'
	size += len(g.Data)

	return
}

// HashTreeRoot ssz hashes the GossipTestData object
func (g *GossipTestData) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(g)
}

// HashTreeRootWith ssz hashes the GossipTestData object with a hasher
func (g *GossipTestData) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'GetData'
	if len(g.Data) > 16777216 {
		err = ssz.ErrBytesLength
		return
	}
	hh.PutBytes(g.Data)

	// Field (1) 'Count'
	hh.PutUint64(g.Count)

	// Field (2) 'Step'
	hh.PutUint64(g.Step)

	hh.Merkleize(indx)
	return
}

// MarshalSSZ ssz marshals the SignedGossipTestData object
func (s *SignedGossipTestData) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(s)
}

// MarshalSSZTo ssz marshals the SignedGossipTestData object to a target array
func (s *SignedGossipTestData) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(52)

	// Offset (0) 'GetData'
	dst = ssz.WriteOffset(dst, offset)
	if s.Data == nil {
		s.Data = new(GossipTestData)
	}
	offset += s.Data.SizeSSZ()

	// Field (1) 'Signature'
	if len(s.Signature) != 48 {
		err = ssz.ErrBytesLength
		return
	}
	dst = append(dst, s.Signature...)

	// Field (0) 'GetData'
	if dst, err = s.Data.MarshalSSZTo(dst); err != nil {
		return
	}

	return
}

// UnmarshalSSZ ssz unmarshals the SignedGossipTestData object
func (s *SignedGossipTestData) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 52 {
		return ssz.ErrSize
	}

	tail := buf
	var o0 uint64

	// Offset (0) 'GetData'
	if o0 = ssz.ReadOffset(buf[0:4]); o0 > size {
		return ssz.ErrOffset
	}

	if o0 < 52 {
		return ssz.ErrInvalidVariableOffset
	}

	// Field (1) 'Signature'
	if cap(s.Signature) == 0 {
		s.Signature = make([]byte, 0, len(buf[4:52]))
	}
	s.Signature = append(s.Signature, buf[4:52]...)

	// Field (0) 'GetData'
	{
		buf = tail[o0:]
		if s.Data == nil {
			s.Data = new(GossipTestData)
		}
		if err = s.Data.UnmarshalSSZ(buf); err != nil {
			return err
		}
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the SignedGossipTestData object
func (s *SignedGossipTestData) SizeSSZ() (size int) {
	size = 52

	// Field (0) 'GetData'
	if s.Data == nil {
		s.Data = new(GossipTestData)
	}
	size += s.Data.SizeSSZ()

	return
}

// HashTreeRoot ssz hashes the SignedGossipTestData object
func (s *SignedGossipTestData) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(s)
}

// HashTreeRootWith ssz hashes the SignedGossipTestData object with a hasher
func (s *SignedGossipTestData) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'GetData'
	if err = s.Data.HashTreeRootWith(hh); err != nil {
		return
	}

	// Field (1) 'Signature'
	if len(s.Signature) != 48 {
		err = ssz.ErrBytesLength
		return
	}
	hh.PutBytes(s.Signature)

	hh.Merkleize(indx)
	return
}
