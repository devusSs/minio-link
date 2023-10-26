# minio-link by devusSs

## What is this for?

This program is intended to be used as a CLI application to upload any file on your system to a [Minio](https://min.io/) instance and then shorten the URL / share link you receive from [Minio](https://min.io/) via [YOURLS](https://yourls.org/) to make it easier and visually nicer to share the file you just uploaded.
You may also download files you have previously uploaded using this program.

## What do you need?

You will need both a working [Minio](https://min.io/) instance and a working [YOURLS](https://yourls.org/) instance. Please also make sure both instances are actually accessible from the public or at least your IP address.

If you have any issues setting up your instances please read the documentation on their sites.

You will also need an operating system which is supported by the program. Currently MacOS, Windows and Linux are supported. Here is a full list:

- MacOS
- Windows 7 (higher versions should also work)
- Linux (requires 'xclip' or 'xsel' to be installed)

## Setup

Simply download an already compiled release file from the [releases](https://github.com/devusSs/minio-link/releases) section and make sure the archive matches your operating system and architecture. Also please take note that only the latest release will be supported fully and different releases may not work anymore or may be removed in the future.

If you do not wish to use a precompiled version of this program head to the [Building](https://github.com/devusSs/minio-link/blob/main/README.md#Building) section.

You will then need to setup an .env file to configure the uploader.
The following .env structure only specifies the recommended parameters. If you are looking for more information / options please check the `internal/config/environment` package.

```env
LINK_MINIO_ENDPOINT=minio.instance.you
LINK_MINIO_ACCESS_KEY=
LINK_MINIO_ACCESS_SECRET=
LINK_MINIO_USE_SSL=true
LINK_YOURLS_ENDPOINT=https://yourlsinstance.you
LINK_YOURLS_SIGNATURE_KEY=
```

## Running

The CLI application provides commands which can be queried via `minio-link --help`.

Using `minio-link help <command>` may also be useful to check configuration options / flags for each command.

The most useful ones for the average user will be the following:
- `update` to upload a file to private or public (default) bucket on your [Minio](https://min.io/) instance and shorten the url via [YOURLS](https://yourls.org/)
- `download` to download a file via it's [YOURLS](https://yourls.org/) url (you may specify a custom download output via flags)
- `update` to update the application automatically if there is a new precompiled release


### Note

This program will automatically copy the final links (either the [Minio](https://min.io/) link if something fails on the [YOURLS](https://yourls.org/) side or the final [YOURLS](https://yourls.org/) shortened link) to your clipboard and may therefor clear any input you have had there before. Please make sure you do not have anything important in your clipboard before using this tool.

## Building

If you want to build the app yourself you will need the following tools:
- The [Go](https://go.dev) programming language to compile the app.
- The [Make](https://www.gnu.org/software/make/) toolchain to make it easier compiling the program.

You can then use the `set_dev.sh` (e.g. `source set_dev.sh`) file to simplify setting up a building / testing environment with a custom version number.

The `Makefile` includes useful functions for building and testing. If do not know how to use the `make` tool please either use a precompiled binary or refrain from using this program.

## Future plans

Check future plans and features in the [roadplan file](https://github.com/devusSs/minio-link/blob/main/roadmap.md).

## Issues

If you have any issues please either [open an issue](https://github.com/devusSs/minio-link/issues) or [send me an e-mail if the issue is urgent](mailto:devuscs@gmail.com). Usually a simple  search on your favourite search engine or reading the (very basic) documentation helps tho.

## Disclaimer

I do not take responsibility for any errors or issues this program might cause for you. Please only use this program if you actually know what it does and you are sure on how to use it.

This program is also not intended to replace any of the awesome and existing projects like [Minio](https://min.io/) or [YOURLS](https://yourls.org/) which are awesome tools I use on a daily basis. Please make sure you are familiar with them before using this tool.

I also do not own claim any copyright to any of the tools or libraries used. Pleae make sure to check out the awesome libraries I used to create this tool in the `go.mod` file.

This program and the repository are a fun / side project, please do not use this tool in production unless you know the risks and are ready to accept and handle potential problems / consequences.