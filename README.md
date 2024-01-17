# Anti Boring Boring Club

- [Token Generator](./gen)
  - The [main program](./gen/cmd/gen/main.go) uses the config file [abbc.yml](./gen/abbc.yml) for the traits names, file paths and probabilities.

- [Mint Contract](./mint)
  - The [smart contract](./mint/contracts/AntiBoringBoringClub.sol) allows 4444 tokens to be minted including a whitelist.

## TODO

- Save metadata json in database
  - Schema
- Server for metadata JSON as an API
- Website to change trait colors
  - Website frontend
    - Design
    - Code
  - Server that generates new image and saves changes to database
- Provenance hash of original tokens

