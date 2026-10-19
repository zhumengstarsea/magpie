cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1294"
  sha256 arm:   "c1ea6718aae9ead00e7d9c1d599ebc9b97a0c04cbda1d31564d9aadfb3f97c7e",
         intel: "71a7f9362eea9ad4b6abda41f4d7b5cb513becfe74d5306933b155428ed2e750"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
