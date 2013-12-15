GO_BUILD_OPTS=-ldflags -s

default: upgorade-vim

clean:
	rm -f upgorade.exe upgorade
	rm -f upgorade-vim.exe upgorade-vim

upgorade: upgorade.go
	go build $(GO_BUILD_OPTS) upgorade.go

upgorade-vim: upgorade-vim.go
	go build $(GO_BUILD_OPTS) upgorade-vim.go
