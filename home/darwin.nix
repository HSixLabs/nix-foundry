{ config, pkgs, lib, users, ... }:

{
  home.packages = with pkgs; [
    m-cli
    mas
  ];

  home.file.".config/iterm2/com.googlecode.iterm2.plist".text = builtins.toJSON {
    "Normal Font" = "MesloLGS NF Regular 12";
    "Terminal Type" = "xterm-256color";
    "Horizontal Spacing" = 1;
    "Vertical Spacing" = 1;
    "Minimum Contrast" = 0;
    "Use Bold Font" = true;
    "Use Bright Bold" = true;
    "Use Italic Font" = true;
    "ASCII Anti Aliased" = true;
    "Non-ASCII Anti Aliased" = true;
    "Use Non-ASCII Font" = false;
    "Ambiguous Double Width" = false;
    "Draw Powerline Glyphs" = true;
    "Only The Default BG Color Uses Transparency" = true;
    "Default Bookmark Guid" = "${users.username}";
    "New Bookmarks" = [
      {
        "Name" = "${users.username}";
        "Guid" = "${users.username}";
        "Custom Directory" = "Recycle";
        "Working Directory" = "${users.homeDirectory}";
      }
    ];
  };

  xdg.enable = true;
} 