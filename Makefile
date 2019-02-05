# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_ADMIN=app/jiacrontab_admin/jiacrontab_admin
BINARY_NODE=app/jiacrontabd/jiacrontabd
BINARY_ADMIN_UNIX=$(BINARY_ADMIN)_unix
BINARY_NODE_UNIX=$(BINARY_NODE)_unix
WORKDIR=./build
ADMINDIR=$(WORKDIR)/jiacrontab_admin
NODEDIR=$(WORKDIR)/jiacrontabd


.PHONY: all build test clean run build-linux build-windows
all: test build
build:
	rm -rf $(WORKDIR)
	mkdir $(WORKDIR)
	mkdir $(ADMINDIR)
	mkdir $(NODEDIR)
	cp app/jiacrontab_admin/jiacrontab_admin.ini $(ADMINDIR)
	cp -r jiacrontab_admin/static $(ADMINDIR)
	cp app/jiacrontabd/jiacrontabd.ini $(NODEDIR)
	$(GOBUILD) -mod=vendor -o $(BINARY_ADMIN) -v ./jiacrontab_admin
	$(GOBUILD) -mod=vendor -o $(BINARY_NODE) -v ./jiacrontabd
	mv $(BINARY_ADMIN) $(ADMINDIR)
	mv $(BINARY_NODE) $(NODEDIR)
test:
	$(GOTEST) -mod=vendor -v ./jiacrontab_admin
	$(GOTEST) -mod=vendor -v ./jiacrontabd
clean:
	rm -f $(BINARY_NODE_UNIX)
	rm -f $(BINARY_ADMIN_UNIX)
	rm -f $(BINARY_ADMIN)
	rm -f $(BINARY_NODE)
	rm -rf $(WORKDIR)
run:
	$(GOBUILD) -mod=vendor -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)


# Cross compilation
build-linux:
	rm -rf $(WORKDIR)
	mkdir $(WORKDIR)
	mkdir $(ADMINDIR)
	mkdir $(NODEDIR)
	cp app/jiacrontab_admin/jiacrontab_admin.ini $(ADMINDIR)
	cp -r jiacrontab_admin/static $(ADMINDIR)
	cp app/jiacrontabd/jiacrontabd.ini $(NODEDIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -mod=vendor -o $(BINARY_ADMIN) -v ./jiacrontab_admin
	GOOS=linux GOARCH=amd64 $(GOBUILD) -mod=vendor -o $(BINARY_NODE) -v ./jiacrontabd
	mv $(BINARY_ADMIN) $(ADMINDIR)
	mv $(BINARY_NODE) $(NODEDIR)

build-windows:
	rm -rf $(WORKDIR)
	mkdir $(WORKDIR)
	mkdir $(ADMINDIR)
	mkdir $(NODEDIR)
	cp app/jiacrontab_admin/jiacrontab_admin.ini $(ADMINDIR)
	cp -r jiacrontab_admin/static $(ADMINDIR)
	cp app/jiacrontabd/jiacrontabd.ini $(NODEDIR)
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC="x86_64-w64-mingw32-gcc -fno-stack-protector -D_FORTIFY_SOURCE=0 -lssp" $(GOBUILD) -mod=vendor -o $(BINARY_ADMIN).exe -v ./jiacrontab_admin
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC="x86_64-w64-mingw32-gcc -fno-stack-protector -D_FORTIFY_SOURCE=0 -lssp" $(GOBUILD) -mod=vendor -o $(BINARY_NODE).exe -v ./jiacrontabd

	mv $(BINARY_ADMIN).exe $(ADMINDIR)
	mv $(BINARY_NODE).exe $(NODEDIR)
