VERSION=0.2

PKGNAME_WIN32=upgorade-vim-win32-$(VERSION)-bin
GO_BUILD_OPTS=-ldflags -s

default: upgorade-vim

clean:
	rm -f upgorade.exe upgorade
	rm -f upgorade-vim.exe upgorade-vim

package: $(PKGNAME_WIN32).zip

upgorade: upgorade.go common/*.go
	go build $(GO_BUILD_OPTS) upgorade.go

upgorade-vim: upgorade-vim.go common/*.go
	go build $(GO_BUILD_OPTS) upgorade-vim.go

$(PKGNAME_WIN32).zip: upgorade-vim
	rm -f $@
	rm -rf $(PKGNAME_WIN32)
	mkdir $(PKGNAME_WIN32)
	cp upgorade-vim.exe $(PKGNAME_WIN32)/
	zip -r9 $(PKGNAME_WIN32).zip $(PKGNAME_WIN32)
	rm -rf $(PKGNAME_WIN32)
