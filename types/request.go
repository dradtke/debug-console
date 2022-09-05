package types

type RequestSender func(name string, args any) (Response, error)
