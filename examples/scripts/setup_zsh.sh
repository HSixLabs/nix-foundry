#!/usr/bin/env bash

rm -rf "$HOME/.oh-my-zsh" || true
RUNZSH=no CHSH=no sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"
rm -rf "${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/themes/powerlevel10k" || true
git clone --depth=1 https://github.com/romkatv/powerlevel10k.git ${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/themes/powerlevel10k
sed -i '' 's/ZSH_THEME="[^"]*"/ZSH_THEME="powerlevel10k\/powerlevel10k"/' ~/.zshrc
echo "# Enable Powerlevel10k instant prompt" > ~/.zshrc.tmp
echo "if [[ -r \"\${XDG_CACHE_HOME:-\$HOME/.cache}/p10k-instant-prompt-\${(%):-%n}.zsh\" ]]; then" >> ~/.zshrc.tmp
echo "  source \"\${XDG_CACHE_HOME:-\$HOME/.cache}/p10k-instant-prompt-\${(%):-%n}.zsh\"" >> ~/.zshrc.tmp
echo "fi" >> ~/.zshrc.tmp
echo "" >> ~/.zshrc.tmp
cat ~/.zshrc >> ~/.zshrc.tmp
echo "" >> ~/.zshrc.tmp
echo "# To customize prompt, run 'p10k configure' or edit ~/.p10k.zsh" >> ~/.zshrc.tmp
echo "[[ ! -f ~/.p10k.zsh ]] || source ~/.p10k.zsh" >> ~/.zshrc.tmp
mv ~/.zshrc.tmp ~/.zshrc
