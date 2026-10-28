# A Nix flake file, mainly intended for developing TinyGo.
# You can download Nix here, for use on your Linux or macOS system:
#   https://nixos.org/download.html
# After you have installed Nix, you can enter the development environment as
# follows:
#
#   nix develop
#
# This drops you into a bash shell, where you can install TinyGo simply using
# the following command:
#
#   go install
#
# That's all! Assuming you've set up your $PATH correctly, you can now use the
# tinygo command as usual:
#
#   tinygo version
#
# But you'll need a bit more to make TinyGo actually able to compile code:
#
#   make llvm-source            # fetch compiler-rt
#   git submodule update --init # fetch lots of other libraries and SVD files
#   make gen-device -j4         # build src/device/*/*.go files
#
# With this, you should have an environment that can compile anything - except
# for the Xtensa architecture (ESP8266/ESP32) because support for that lives in
# a separate LLVM fork.
#
# You can also do many other things from this environment. Building and flashing
# should work as you're used to: it's not a VM or container so there are no
# access restrictions and you're running in the same host environment - just
# with a slightly different set of tools available.
{
  inputs = {
    # unstable, for go_1_27. The lock fixes the version to make it reproducible.
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };
  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      with pkgs;
      {
        # shortcut: only the sources needed for the wasm targets are installed,
        # add the other submodules here to support more targets.
        packages.default = buildGo127Module {
          pname = "tinygo";
          version = "0.44.0-dev";
          src = self;
          vendorHash = "sha256-YLoanj1gV4NNXQu/d0B/nl9lImh8m4wI/Tc9hpRdgQg=";

          subPackages = [ "." ];
          tags = [ "llvm20" ];
          ldflags = [
            # tinygo keys its cache on the build id, buildGoModule would leave it empty
            "-buildid=${self.rev or self.dirtyRev or "dirty"}"
            "-X github.com/tinygo-org/tinygo/goenv.TINYGOROOT=${placeholder "out"}/share/tinygo"
            "-X github.com/tinygo-org/tinygo/goenv.clangResourceDir=${llvmPackages_20.clang.cc.lib}/lib/clang/20"
          ];

          nativeBuildInputs = [ makeWrapper ];
          buildInputs = [
            llvmPackages_20.llvm
            llvmPackages_20.libclang
          ];
          doCheck = false;
          # the wrapper puts go on PATH
          allowGoReference = true;

          # flake sources come without submodules, keep these revisions in sync with them
          postInstall = ''
            root=$out/share/tinygo
            mkdir -p $root/lib
            cp -r src targets $root/
            rm -rf $root/src/net
            cp -r ${fetchFromGitHub {
              owner = "tinygo-org";
              repo = "net";
              rev = "70037cf71f1ae14175781e162c9c76d1ac730ece";
              hash = "sha256-k1gGbXm4oO5lh5pP0I8rI1XZQNDG2fBmvvgdHsVrG3w=";
            }} $root/src/net
            cp -r ${fetchFromGitHub {
              owner = "WebAssembly";
              repo = "wasi-libc";
              rev = "1dfe5c302d1c5ab621f7abf04620fae92700fd22";
              hash = "sha256-Km0HPHXFfN373zW+jdND7vB4/R589uXPe5KcvmAN+nw=";
            }} $root/lib/wasi-libc
            cp -r ${llvmPackages_20.compiler-rt.src}/compiler-rt/lib/builtins $root/lib/compiler-rt-builtins
            chmod -R u+w $root

            wrapProgram $out/bin/tinygo --prefix PATH : ${lib.makeBinPath [
              go_1_27
              llvmPackages_20.clang.cc
              llvmPackages_20.lld
              binaryen
            ]}
          '';
        };

        devShells.default = mkShell {
          buildInputs = [
            # These dependencies are required for building tinygo (go install).
            go_1_27
            llvmPackages_20.llvm
            llvmPackages_20.libclang
            # Additional dependencies needed at runtime, for building and/or
            # flashing.
            llvmPackages_20.lld
            avrdude
            binaryen
            # Additional dependencies needed for on-chip debugging.
            # These tools are rather big (especially GDB) and not frequently
            # used, so are commented out. On-chip debugging is still possible if
            # these tools are available in the host environment.
            #gdb
            #openocd
          ];
          shellHook= ''
            # Make `make smoketest` work (the default is `md5`, while Nix only
            # has `md5sum`).
            export MD5SUM=md5sum

            # Ugly hack to make the Clang resources directory available.
            export GOFLAGS="\"-ldflags=-X github.com/tinygo-org/tinygo/goenv.clangResourceDir=${llvmPackages_20.clang.cc.lib}/lib/clang/20\" -tags=llvm20"
          '';
        };
      }
    );
}
