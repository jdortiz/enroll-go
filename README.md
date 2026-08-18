# enroll: Building a Verified Information Network with Vonage Network APIs

## This Repo

This repository contains the Go code explained in my presentation "The Human Network: Using Cellular Intelligence to Vet
Who Gets In".

I have included a brief [description of the architecture](./ARCHITECTURE.md).

## Use case

KommKorp is a news agency. These days, with so much disinformation, it is hard to find the best local sources for
breaking news everywhere. KommKorp has correspondents all over the world and a network of informants that bring the most
amazing scoops.

WARNING: If you don't work for a news agency, bear with me. What I am going to explain may also be helpful to your
company.

The process of bringing new informants into the network is highly sensitive and must be conducted remotely. So KommKorp
has decided to enroll the informants using their trust in the correspondents.

I chose this use case because it can benefit from several features of the Network APIs and allows us to reason about
which insights can be useful for the business logic.

This code implements the core components of the enrollment process in a server-side application written in Go, using
Vonage Network APIs to simplify and enrich our process logic.

If you want to try this code, you will need:

- A recent version of the Go development environment. I am using 1.26.4, which is the latest version at the moment.
- A Vonage developer account to be able to use the APIs. You can get one from https://dashboard.vonage.com Create an
  application and enable “Verify” and “Network Registry” using the playground.

## Network APIs

This code uses the Verify (v2) and Identity Insights APIs. The correspondent identity is verified using 2 factors: a
traditional password (something you know) and a code sent to the correspondent’s phone (something you have), provided by
the Verify API. Information about the prospective informant is verified using the Identity Insights API, allowing
KommKorp to filter out candidates it cannot trust. An informant is bound to a country, but some might use a phone on a
carrier in a different country for confidentiality reasons. So, I decided to use two pieces of data from Identity
Insights: the format, to check whether the number format is the right one for the informant’s country or if it belongs
to a different country, it is roaming in their country.

You can consider using other identity insights, but some of them are harder to apply in this case. For example, locating
the informant’s device would check whether it is within a circle of a few meters around the position, but that is too
narrow for this use case.

## Setup instructions

Create a `.env` file with your secret at the root of the project. It should be similar to the following one:
```
JWT_SECRET="-----BEGIN PRIVATE KEY-----
MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQDnZNhQpfjz8uJE
X2+PBNvO/yePbPikjDTn7wDAhuEoYToMP2/WTXoz98dlBkxi7TKXLfsBFt9Mgr5K
... Omitted ...
GMaDpeS3QlsX+1hawmvuM8T4Sjr5qoep+c4fPZD+LL9o/jdT3r6new97klpVqLtB
WsdO0KoAjLxslpX8Eso2DA==
-----END PRIVATE KEY-----"
JWT_PUBLIC_KEY="-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA52TYUKX48/LiRF9vjwTb
... Omitted ...
sUsL/LrtSKV/5a67zP1j0NMjO1PBlUEE8+sJh0v6taJv2Mj+xJaHc6i4/C8L0n3q
2QIDAQAB
-----END PUBLIC KEY-----"
VONAGE_APP_ID="12345678-abcd-efgh-9012-1a2b3c4d5e6f"
```

The associated public key must be uploaded to the application defined in the dashboard.

WARNING: `.env` files are in .gitignore so they won't be committed to the repository accidentally.

Once your `.env` file has been created, you can execute the project with:
```shell
GO_LOG=debug go run cmd/enroll/main.go
```

It will launch an HTTP service http://127.0.0.1:8080 and the `GO_LOG` environment variable, will ensure that all logging
messages are displayed.

## Assumptions, issues, and limitations

Commits to this codebase have been made to show progressive implementation of the application.

The users are stored in a simplified `UserInMemoryRepo`. Actual code should use an actual database, but the
`UserRepository` interface simplifies this change using the [Dependency Inversion
Principle](https://en.wikipedia.org/wiki/Dependency_inversion_principle).

HTTP service is not encrypted. `ListenAndServeTLS` should be used in `(*App) Start()` providing the certificate and
private key. Otherwise, login requests are sent in the clear.

Secrets are passed via environment variables set from a `.env` file. That means that the private key is in plain text in
memory. A secure implementation would require a TPM or another device that implements PKCS#11 or similar. Also, it would
be better to check that all the secrets are in the file at the beginning, rather than failing when they are needed.

No proper error handling has been implemented. For example, if password hashing fails during `User` creation, the
application will exit with an error.

Proper logging must also be added to the application. The existing structured logging are meant for demoing and
debugging purposes.


## Time spent in this project



## License

This project is licensed under the terms of the [Apache license 2.0](./LICENSE.txt).

## Author

Jorge D. Ortiz Fuentes, 2026

## Resources

To learn more about Jorge's work you have all these fantastic resources:

- [🖋️ Jorge's Blog](https://jorgeortiz.dev/)
- [🙋🏻‍♂️ Jorge's LinkedIn](https://www.linkedin.com/in/jorgeortiz/)
- [🙋🏻‍♂️ Jorge's Reddit](https://www.reddit.com/user/jorgedortiz/)
- [🙋🏻‍♂️ Jorge's Bluesky](https://bsky.app/profile/jdortiz.bsky.social)
- [🙋🏻‍♂️ Jorge's Mastodon](https://fosstodon.org/@jdortiz)
- [🙋🏻‍♂️ Jorge's X](https://x.com/jdortiz)
