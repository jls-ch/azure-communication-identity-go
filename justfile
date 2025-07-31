format:
    golangci-lint fmt
lint:
    golangci-lint run

# to publish, tag commit(vX.Y.Z) and push the tag to origin, ensure SEMVER!!!
check-published VERSION:
    GOPROXY=proxy.golang.org go list -m github.com/jls-ch/azure-communication-identity-go@{{VERSION}}
