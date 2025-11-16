package events

import (
	"encoding/json"
	"fmt"
)

func Parse(data []byte) (any, error) {
	var base BaseEvent
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, err
	}

	switch base.Type {
	case "sys":
		var e Sys
		err := json.Unmarshal(data, &e)
		return e, err
	case "broadcast":
		var e Broadcast
		err := json.Unmarshal(data, &e)
		return e, err
	case "auth":
		var e Auth
		err := json.Unmarshal(data, &e)
		return e, err
	case "msg":
		var e Msg
		err := json.Unmarshal(data, &e)
		return e, err
	case "voice":
		var e Voice
		err := json.Unmarshal(data, &e)
		return e, err
	case "cmd":
		var e Cmd
		err := json.Unmarshal(data, &e)
		return e, err
	case "usr_list":
		var e UsrList
		err := json.Unmarshal(data, &e)
		return e, err
	case "chan_list":
		var e ChanList
		err := json.Unmarshal(data, &e)
		return e, err
	default:
		return nil, fmt.Errorf("unknown event type: %s", base.Type)
	}
}

func panicMsg(e any, err error) string {
	return fmt.Sprintf("failed to serialise event\nevent: %+v\nerror: %v", e, err)
}

func (e Sys) Serialise() string {
	data, err := json.Marshal(e)
	if err != nil {
		panic(panicMsg(e, err))
	}
	return string(data)
}

func (e Broadcast) Serialise() string {
	data, err := json.Marshal(e)
	if err != nil {
		panic(panicMsg(e, err))
	}
	return string(data)
}

func (e Voice) Serialise() string {
	data, err := json.Marshal(e)
	if err != nil {
		panic(panicMsg(e, err))
	}
	return string(data)
}

func (e UsrList) Serialise() string {
	data, err := json.Marshal(e)
	if err != nil {
		panic(panicMsg(e, err))
	}
	return string(data)
}

func (e ChanList) Serialise() string {
	data, err := json.Marshal(e)
	if err != nil {
		panic(panicMsg(e, err))
	}
	return string(data)
}
