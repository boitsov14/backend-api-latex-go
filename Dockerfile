FROM alpine:latest AS installer
RUN apk --no-cache add perl tar wget
WORKDIR /install-tl-unx
RUN wget https://mirror.ctan.org/systems/texlive/tlnet/install-tl-unx.tar.gz
RUN tar xvzf ./install-tl-unx.tar.gz --strip-components=1
COPY texlive.profile .
RUN ./install-tl --profile=texlive.profile
RUN mkdir -p /usr/local/texlive/bin
RUN ln -sf /usr/local/texlive/*/bin/* /usr/local/texlive/bin/texlive
ENV PATH=/usr/local/texlive/bin/texlive:$PATH
RUN tlmgr install bussproofs cbfonts-fd dvisvgm ebproof forest greek-fontenc lplfitch preview standalone varwidth

# FROM scratch
# COPY --from=installer /usr/local/texlive /usr/local/texlive

FROM alpine AS build
# Install Ghostscript
RUN apk --no-cache add ghostscript

FROM cgr.dev/chainguard/static
# Install TeX
# COPY --from=boitsov14/minimal-prooftree-latex /usr/local/texlive /usr/local/texlive
COPY --from=installer /usr/local/texlive /usr/local/texlive
ENV PATH=/usr/local/texlive/bin/texlive:$PATH
# Copy Ghostscript
COPY --from=build /usr/bin/gs /usr/bin/gs
# Copy libc
COPY --from=build /lib /lib
COPY --from=build /usr/lib /usr/lib
# Serve the binary
WORKDIR /app
EXPOSE 3001
COPY bin/main .
ENTRYPOINT ["./main"]
