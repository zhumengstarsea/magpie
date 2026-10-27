cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1661"
  sha256 arm:   "8cfc373fd1456da90f7f31224c0b16cee4fefc82303e184b81886a4223c64936",
         intel: "99f5c00510aec565d4106fc4753fa4a7b01327db3ae989e36d42070d10a24099"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
