package protocol

func NewErrorMsg(err string) Msg {
	return Msg{
		Type: "error",
		Msg:  err,
	}
}
