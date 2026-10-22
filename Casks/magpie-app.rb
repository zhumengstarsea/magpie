cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1318"
  sha256 arm:   "5fa45f543d8e0a0b0b34d2cebec3da1b3374b7705750f24d0a1847f247be2acd",
         intel: "91e314f43139bf244f0918baced122d9ad8e8f9b3e925c55ba5b105b9111720b"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
