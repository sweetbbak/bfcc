default:
    just bfcc

bfcc:
    go build -ldflags "-s -w" -o bfcc ./cmd/bfcc

bftui:
    go build -ldflags "-s -w" -o bftui ./cmd/bftui
