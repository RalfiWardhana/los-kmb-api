# Platform Library

Platform library module is used as a tools for internal company Kreditplus.

---

<details>
<summary><b>View table of contents</b></summary><br/>

- [Platform Library](#platform-library)
- [Getting Started](#getting-started)
  - [Setup](#setup)
    - [Config Git](#config-git)
      - [1. First setup git configuration](#1-first-setup-git-configuration)
      - [2. Second by passing go mod proxy site](#2-second-by-passing-go-mod-proxy-site)
      - [3. Third passing repository credentials to Go module during the build](#3-third-passing-repository-credentials-to-go-module-during-the-build)
      - [4. Fourth set up your editor code](#4-fourth-set-up-your-editor-code)
  - [Install](#install)
  - [Examples](#examples)
    - [Masking data](#masking-data)
    - [OCR](#ocr)
    - [SSO](#sso)
    - [Notification](#notification)
  - [Example Code](#example-code)
  - [Contributing](#contributing)
  - [License](#license)

</details>

---

# Getting Started

Here some preparations that you might need before using this package.

## Setup 

Here how to use private repositories with Go modules:

### Config Git

#### 1. First setup git configuration 
```bash
$ git config --global user.name "username"
$ git config --global user.email "email.company"
```

#### 2. Second by passing go mod proxy site
```bash
$ go env -w GOPRIVATE=github.com/KB-FMF/*
```

#### 3. Third passing repository credentials to Go module during the build

> Note: Auth token must be URL encoded.

Click [here](https://github.com/settings/tokens) to go to GitHub access tokens. Create one with appropriate permissions. Create one new token or use an existing one.

![github_token](./github_token.png)

After generation of token execute the following command.

```bash
$ git config --global url."https://${username}:${access_token}@github.com".insteadOf / "https://github.com"
```

#### 4. Fourth set up your editor code 

If you're using Goland editor code you can set modules proxy. go to Settings > Go Modules > Environtment > Choose GOPRIVATE and put there > Apply and OK.

![goland](./gomod_goland.png)

It can enable go modules integration on private repos.

If using Visual studio code don't worry, that might now be supported, as a consequence of gopls v0.5.1.

> Even if you don't have multiple modules, a go.mod in a sub-folder (instead of the root folder of your project) will better managed (if you activate the `gopls.experimentalWorkspaceModule` setting).


## Install

```bash
go get -v github.com/KB-FMF/platform-library@0.1
```

## Examples

The following are simple examples:

### Masking data

Example masking data [see](./example/maskingdata/README.md)

### OCR

Example ocr [see](./example/ocr/README.md)

### SSO

Example sso [see](./example/sso/README.md)

### Auth Application

Example auth application [see](./example/auth/README.md)

### Notification

Example notification [see](./example/notification/README.md)

### OTP

Example otp [see](./example/otp/README.md)


## Example Code
For more complex examples, refer to the [example/](/example) directory.

You can find the sample code in [here](/example)

## Contributing

No rules for now. Feel free to add issue first and optionally submit a PR. Cheers

## License

MIT. Copyright 2022 [Kreditplus](./LICENSE)
