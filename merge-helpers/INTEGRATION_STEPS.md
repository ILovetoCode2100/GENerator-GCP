# Manual Integration Steps

## 1. Client.go Integration

Compare the following files and merge Version B's methods into Version A:
- Version A: `/pkg/virtuoso/client.go`
- Version B reference: `/merge-helpers/client-version-b.go`

Key methods to add from Version B:
- All CreateStep* methods for the new commands
- Helper methods like createStepWithCustomBody()

## 2. Main.go Command Registration

Update Version A's `/src/cmd/main.go` to include Version B's command registrations.
Reference: `/merge-helpers/main-version-b.go`

Add the following command registrations to the init() function:
```go
// Cookie management
rootCmd.AddCommand(newCreateStepCookieCreateCmd())
rootCmd.AddCommand(newCreateStepCookieWipeAllCmd())

// ... (see main-version-b.go for complete list)
```

## 3. Module Updates

Ensure go.mod includes necessary dependencies:
```
github.com/go-resty/resty/v2 v2.11.0
gopkg.in/yaml.v2 v2.4.0
```

## 4. Testing

After integration:
1. Run `go build` in Version A directory
2. Execute the test scripts to verify functionality
3. Test both old (project management) and new (enhanced steps) features

