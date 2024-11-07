# TODO: adjust the rocksdb stuff and Evmos mentions
{ lib
, buildGoApplication
, buildPackages
, fetchFromGitHub
, stdenv
, rev ? "dirty"
# TODO: remove the rocksdb stuff, that's only used on Evmos
, rocksdb
, static ? stdenv.hostPlatform.isStatic
, dbBackend ? "goleveldb"
}:
let
  version = if dbBackend == "rocksdb" then "latest-rocksdb" else "latest";
  pname = "osd";
  tags = [ "ledger" "netgo" ] ++ lib.optionals (dbBackend == "rocksdb") [ "rocksdb" "grocksdb_clean_link" ];
  ldflags = lib.concatStringsSep "\n" ([
    "-X github.com/cosmos/cosmos-sdk/version.Name=evmOS"
    "-X github.com/cosmos/cosmos-sdk/version.AppName=${pname}"
    "-X github.com/cosmos/cosmos-sdk/version.Version=${version}"
    "-X github.com/cosmos/cosmos-sdk/version.BuildTags=${lib.concatStringsSep "," tags}"
    "-X github.com/cosmos/cosmos-sdk/version.Commit=${rev}"
    "-X github.com/cosmos/cosmos-sdk/types.DBBackend=${dbBackend}"
  ]);
  buildInputs = lib.optionals (dbBackend == "rocksdb") [ rocksdb ];
  # use a newer version of nixpkgs to get go_1_22
  # We're not updating this on the whole setup because breaks other stuff
  # but we can import the needed packages from the newer version
  nixpkgsUrl = "https://github.com/NixOS/nixpkgs/archive/master.tar.gz";
  nixpkgs = import (fetchTarball nixpkgsUrl) {};
  # the go_1_22 nixpkgs is v1.22.6
  # but we need the v1.22.8. 
  # This overrides the pkg to use
  # the v1.22.8 version  
  go_1_22 = nixpkgs.pkgs.go_1_22.overrideAttrs {
    pname = "golang";
    version = "go1.22.8";
    src = fetchFromGitHub {
      owner = "golang";
      repo = "go";
      rev = "aeccd613c896d39f582036aa52917c85ecf0b0c0";
      sha256 = "sha256-N3uG+FLMgThIAr1aDJSq+X+VKCz8dw6az35um3Mr3D0=";

    };
  };
in
buildGoApplication rec {
  inherit pname version buildInputs tags ldflags;
  go = go_1_22;
  src = ./.;
  modules = ./example_chain/gomod2nix.toml;
  doCheck = false;
  pwd = src; # needed to support replace
  subPackages = [ "example_chain/cmd/osd" ];
  CGO_ENABLED = "1";

  postFixup = if dbBackend == "rocksdb" then
    ''
      # Rename the binary from osd to osd-rocksdb
      mv $out/bin/osd $out/bin/osd-rocksdb
    '' else '''';

  meta = with lib; {
    description = "evmOS is a plug-and-play solution that adds EVM compatibility and customizability to your chain!";
    homepage = "https://github.com/evmos/os";
    license = licenses.asl20;
    mainProgram = "osd";
  };
}
