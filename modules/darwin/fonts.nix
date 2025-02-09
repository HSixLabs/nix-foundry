{ config, pkgs, lib, ... }:

{
  fonts = {
    packages = with pkgs; [
      (stdenv.mkDerivation {
        name = "meslo-lgs-nf";
        src = fetchurl {
          url = "https://github.com/romkatv/powerlevel10k-media/raw/master/MesloLGS%20NF%20Regular.ttf";
          sha256 = "2XlGGG6X+NfAE56Jg6v0Ch0tCGkk8sXb8cKb2PLG5X0=";
        };
        dontUnpack = true;
        installPhase = ''
          mkdir -p $out/share/fonts/truetype
          cp $src $out/share/fonts/truetype/MesloLGS-NF-Regular.ttf
        '';
      })
    ];
  };
}