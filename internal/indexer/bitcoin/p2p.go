package bitcoin

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"net"
	"sync"
	"time"
)

const MagicBytes = 0xdab5bffa //identifies bitcoin regtest network

type P2PCaller struct {
	mu           sync.Mutex
	conn         net.Conn
	rpc          *RPCClient ///for hashing block and block count, had not found alternative yet in p2p
	blockFetches map[string]*fetchState
}

type fetchState struct {
	done  chan struct{}
	err   error
	block *BlockHeader
	txs   []Transaction
}

func NewP2PCaller(address string, rpc *RPCClient) (*P2PCaller, error) {
	conn, err := net.DialTimeout("tcp", address, 10*time.Second)

	if address == "" {
		return nil, fmt.Errorf("p2p address is empty")
	}

	if err != nil {
		return nil, fmt.Errorf("p2p connection failed: %w", err)
	}

	p := &P2PCaller{
		conn:         conn,
		rpc:          rpc,
		blockFetches: make(map[string]*fetchState),
	}

	go p.listen()

	if err := p.Handshake(); err != nil {
		return nil, fmt.Errorf("handshake failed: %w", err)
	}

	return p, nil
}

func (p *P2PCaller) listen() {
	fmt.Println("P2P listener started, waiting for headers...")
	defer p.conn.Close()
	for {
		header := make([]byte, 24)
		if _, err := io.ReadFull(p.conn, header); err != nil {
			return
		}

		magic := binary.LittleEndian.Uint32(header[0:4])
		if magic != MagicBytes {
			continue
		}

		command := string(bytes.TrimRight(header[4:16], "\x00"))
		length := binary.LittleEndian.Uint32(header[16:20])

		payload := make([]byte, length)
		if _, err := io.ReadFull(p.conn, payload); err != nil {
			return
		}

		switch command {
		case "version":
			p.sendMessage("verack", nil)
		case "block":
			p.handleBlockMessage(payload)
		case "ping":
			p.sendMessage("pong", payload)
		}
	}
}

func (p *P2PCaller) Handshake() error {
	payload := make([]byte, 85)
	binary.LittleEndian.PutUint32(payload[0:4], 70015)
	binary.LittleEndian.PutUint64(payload[4:12], 1)
	binary.LittleEndian.PutUint64(payload[12:20], uint64(time.Now().Unix()))

	if err := p.sendMessage("version", payload); err != nil {
		return err
	}

	time.Sleep(500 * time.Millisecond)
	return nil
}

func (p *P2PCaller) GetBlock(hash string) ([]Transaction, error) {
	fs, err := p.requestBlock(hash)
	if err != nil {
		return nil, err
	}

	select {
	case <-fs.done:
		return fs.txs, fs.err
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("timeout waiting for block %s", hash)
	}
}

func (p *P2PCaller) requestBlock(hash string) (*fetchState, error) {
	p.mu.Lock()
	if fs, ok := p.blockFetches[hash]; ok {
		p.mu.Unlock()
		return fs, nil
	}

	fs := &fetchState{done: make(chan struct{})}
	p.blockFetches[hash] = fs
	p.mu.Unlock()

	hashBytes, _ := hex.DecodeString(hash)
	for i, j := 0, len(hashBytes)-1; i < j; i, j = i+1, j-1 {
		hashBytes[i], hashBytes[j] = hashBytes[j], hashBytes[i]
	}

	var payload bytes.Buffer
	payload.WriteByte(0x01)

	binary.Write(&payload, binary.LittleEndian, uint32(2))
	payload.Write(hashBytes)

	return fs, p.sendMessage("getdata", payload.Bytes())
}

func (p *P2PCaller) handleBlockMessage(payload []byte) {
	header := payload[:80]
	if len(payload) < 80 {
		return
	}

	hashRaw := doubleSha256(header)
	for i, j := 0, len(hashRaw)-1; i < j; i, j = i+1, j-1 {
		hashRaw[i], hashRaw[j] = hashRaw[j], hashRaw[i]
	}
	blockHash := hex.EncodeToString(hashRaw[:])

	reader := bytes.NewReader(payload[80:])
	txCount, _ := readVarInt(reader)

	var txs []Transaction
	for range txCount {
		tx, _ := parseTransaction(reader)
		txs = append(txs, tx)
	}

	p.mu.Lock()
	fs, ok := p.blockFetches[blockHash]
	if ok {
		fs.txs = txs
		fs.err = nil
		close(fs.done)
		delete(p.blockFetches, blockHash)
	}
	p.mu.Unlock()
}

func (p *P2PCaller) sendMessage(command string, payload []byte) error {
	header := make([]byte, 24)
	binary.LittleEndian.PutUint32(header[0:4], MagicBytes)
	copy(header[4:16], command)
	binary.LittleEndian.PutUint32(header[16:20], uint32(len(payload)))
	sum := doubleSha256(payload)
	copy(header[20:24], sum[:4])

	_, err := p.conn.Write(append(header, payload...))
	return err
}

func parseTransaction(r *bytes.Reader) (Transaction, error) {
	var tx Transaction
	var buf bytes.Buffer
	tr := io.TeeReader(r, &buf)

	binary.Read(tr, binary.LittleEndian, &tx.Version)

	marker := make([]byte, 2)
	r.Read(marker)
	isSegWit := marker[0] == 0x00 && marker[1] == 0x01
	if !isSegWit {
		r.Seek(-2, io.SeekCurrent)
	}

	inCount, _ := readVarInt(tr)
	for i := uint64(0); i < inCount; i++ {
		var in TxInput
		prevHash := make([]byte, 32)
		io.ReadFull(tr, prevHash)
		for i, j := 0, len(prevHash)-1; i < j; i, j = i+1, j-1 {
			prevHash[i], prevHash[j] = prevHash[j], prevHash[i]
		}
		in.PrevTxID = hex.EncodeToString(prevHash)

		var vout uint32
		binary.Read(tr, binary.LittleEndian, &vout)
		in.Vout = int64(vout)

		scriptLen, _ := readVarInt(tr)
		script := make([]byte, scriptLen)
		io.ReadFull(tr, script)

		r.Seek(4, io.SeekCurrent)

		tx.Inputs = append(tx.Inputs, in)
	}

	outCount, _ := readVarInt(tr)
	for i := uint64(0); i < outCount; i++ {
		var out TxOutput
		var satoshis int64
		binary.Read(tr, binary.LittleEndian, &satoshis)
		out.Value = float64(satoshis) / 1e8

		scriptLen, _ := readVarInt(tr)
		script := make([]byte, scriptLen)
		io.ReadFull(tr, script)

		out.Address = extractAddress(script)
		out.Vout = int64(i)

		tx.Outputs = append(tx.Outputs, out)

		fmt.Printf("DEBUG: Found output with address: %s, value: %f\n", out.Address, out.Value)
	}

	var locktime uint32
	binary.Read(tr, binary.LittleEndian, &locktime)
	tx.Locktime = int64(locktime)

	hash := doubleSha256(buf.Bytes())
	for i, j := 0, len(hash)-1; i < j; i, j = i+1, j-1 {
		hash[i], hash[j] = hash[j], hash[i]
	}
	tx.TxID = hex.EncodeToString(hash[:])

	return tx, nil
}

func readVarInt(r io.Reader) (uint64, error) {
	var b byte
	binary.Read(r, binary.LittleEndian, &b)
	if b < 0xfd {
		return uint64(b), nil
	}
	if b == 0xfd {
		var v uint16
		binary.Read(r, binary.LittleEndian, &v)
		return uint64(v), nil
	}
	var v uint32
	binary.Read(r, binary.LittleEndian, &v)
	return uint64(v), nil
}

func doubleSha256(b []byte) [32]byte {
	h := sha256.Sum256(b)
	return sha256.Sum256(h[:])
}

func extractAddress(script []byte) string {
	if len(script) == 0 {
		return "unknown"
	}

	if len(script) == 25 && script[0] == 0x76 && script[1] == 0xa9 && script[24] == 0xac {
		return base58CheckEncode(0x6f, script[3:23])
	}

	if len(script) == 23 && script[0] == 0xa9 && script[22] == 0x87 {
		return base58CheckEncode(0xc4, script[2:22])
	}

	if len(script) >= 4 && script[0] == 0x00 && (script[1] == 20 || script[1] == 32) {
		addr, _ := encodeBech32("bcrt", 0, script[2:])
		return addr
	}

	if script[0] == 0x6a {
		return "OP_RETURN"
	}

	return "unknown_script"
}

func (p *P2PCaller) Call(m string, params []any, res any) error {
	return fmt.Errorf("p2p caller does not support raw rpc call: %s", m)
}

func (p *P2PCaller) GetBlockHash(h int64) (string, error) {
	var hash string
	err := p.rpc.Call("getblockhash", []any{h}, &hash)
	return hash, err
}

func (p *P2PCaller) GetBlockHeader(hash string) (*BlockHeader, error) {
	return p.rpc.GetBlockHeader(hash)
}

func (p *P2PCaller) GetTxOutProof(tx, bh string) ([]byte, error) {
	return nil, fmt.Errorf("merkle proof not supported in p2p mode")
}

var bech32Charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

func hrpExpand(hrp string) []byte {
	var ret []byte
	for _, c := range hrp {
		ret = append(ret, byte(c>>5))
	}
	ret = append(ret, 0)
	for _, c := range hrp {
		ret = append(ret, byte(c&31))
	}
	return ret
}

func bech32Polymod(values []byte) uint32 {
	g := []uint32{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}
	chk := uint32(1)
	for _, v := range values {
		b := chk >> 25
		chk = ((chk & 0x1ffffff) << 5) ^ uint32(v)
		for i := 0; i < 5; i++ {
			if ((b >> i) & 1) == 1 {
				chk ^= g[i]
			}
		}
	}
	return chk
}

func encodeBech32(hrp string, version byte, prog []byte) (string, error) {
	data := []byte{version}
	data = append(data, convertBits(prog, 8, 5, true)...)
	checksum := bech32Checksum(hrp, data)

	all := append(data, checksum...)
	var result []byte
	for _, v := range all {
		if int(v) >= len(bech32Charset) {
			return "", fmt.Errorf("invalid value for bech32: %d", v)
		}
		result = append(result, bech32Charset[v])
	}
	return hrp + "1" + string(result), nil
}

var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

func base58Encode(input []byte) string {
	num := new(big.Int).SetBytes(input)
	var encoded []byte
	base := big.NewInt(58)
	zero := big.NewInt(0)
	mod := new(big.Int)

	for num.Cmp(zero) > 0 {
		num.DivMod(num, base, mod)
		encoded = append([]byte{b58Alphabet[mod.Int64()]}, encoded...)
	}

	for _, b := range input {
		if b != 0 {
			break
		}
		encoded = append([]byte{b58Alphabet[0]}, encoded...)
	}

	return string(encoded)
}

func base58CheckEncode(version byte, payload []byte) string {
	data := append([]byte{version}, payload...)
	hash := doubleSha256(data)
	checksum := hash[:4]
	return base58Encode(append(data, checksum...))
}

func convertBits(data []byte, frombits, tobits uint8, pad bool) []byte {
	var acc uint32 = 0
	var bits uint8 = 0
	var ret []byte
	maxv := uint32(1<<tobits - 1)
	for _, value := range data {
		acc = (acc << frombits) | uint32(value)
		bits += frombits
		for bits >= tobits {
			bits -= tobits
			ret = append(ret, byte((acc>>bits)&maxv))
		}
	}
	if pad {
		if bits > 0 {
			ret = append(ret, byte((acc<<(tobits-bits))&maxv))
		}
	}
	return ret
}

func bech32Checksum(hrp string, data []byte) []byte {
	integers := append(hrpExpand(hrp), data...)
	integers = append(integers, []byte{0, 0, 0, 0, 0, 0}...)
	polymod := bech32Polymod(integers) ^ 1
	var res []byte
	for i := 0; i < 6; i++ {
		res = append(res, byte((polymod>>uint(5*(5-i)))&31))
	}
	return res
}
