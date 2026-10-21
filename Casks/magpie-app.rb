cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1311"
  sha256 arm:   "85048b58c9e35844056f4fea0d8ca74912b2412545c7c1af50d1af6da7c138be",
         intel: "f4b61d5b7a5fcfc9ce8b32b53fb3ecb58c728299b0992716485fe845a7db941c"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
