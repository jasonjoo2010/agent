package transport

import (
	"bufio"
	"encoding/json"

	"github.com/projecteru2/agent/types"
)

// TransportInJson transports logs to writer in json format and separated by new line, return nil when success
func transportInJson(wr *bufio.Writer, logArr []*types.Log, packetSizeLimit int) (err error) {
	for i := 0; i < len(logArr); i++ {
		data, err := json.Marshal(logArr[i])
		if err != nil {
			// ignore
			continue
		}
		if packetSizeLimit > 0 && wr.Buffered()+len(data) > packetSizeLimit {
			// avoid to splitted packet
			wr.Flush()
		}
		_, err = wr.Write(data)
		if err != nil {
			// io error
			return err
		}
		wr.WriteByte('\n')
	}
	return wr.Flush()
}
