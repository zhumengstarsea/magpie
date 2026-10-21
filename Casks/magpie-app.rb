cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1307"
  sha256 arm:   "6c9f9d049729c3c42951058ac59a49716ab24120014d05bfcec4871143be732c",
         intel: "162be1f018707beac118dc262a12ee4cc110ffb49bd303229365a2d79de903c1"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
