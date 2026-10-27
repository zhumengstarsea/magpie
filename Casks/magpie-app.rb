cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1774"
  sha256 arm:   "3b4de0b42bb7573146dbedf41e4faaf8b3136a590f5e3d5758e08b087d93bb64",
         intel: "b9b28b1835efeaeeeab590082eedab264a40e3728fca157edeca8e2a5739523a"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
