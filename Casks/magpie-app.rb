cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1587"
  sha256 arm:   "2c9c12dea32a14517ba261811a49dc2748df20c8b74b736ee0f3319ddecd2dbe",
         intel: "3b022982d40fb480a00ef628bd6018ef78ebfe1e718b4afef0b07bbfe091f6e6"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
