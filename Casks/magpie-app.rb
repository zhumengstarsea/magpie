cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1957"
  sha256 arm:   "8eb6fbba0d7ff83125b155a753b43f97e481357de0b25416686fbfc95ee39d3d",
         intel: "630ead50735e9fbaac6527257749a368020f918f98b1881d16c0810f631293f2"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
