cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1298"
  sha256 arm:   "c1a5c7059f528264d1ad8cdf9d6f6d4542c14274bed8efa05d229399a125e2b9",
         intel: "6268f6bd999cb27918be119800a65bc56c880e40474c168215005fa9940025b8"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
