GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean

build: 
				$(GOBUILD)

test:	clean build
				./test.sh

clean:
				$(GOCLEAN)
				rm -f tmp*