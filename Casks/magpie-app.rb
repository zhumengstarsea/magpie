cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1545"
  sha256 arm:   "85981d926f6d932f6779276dc55099e18eb6a73b5fa837227b8d6807ceeb77f9",
         intel: "786f95a9ab3fa76dbc2aee15534da4c92920e7585dfe10e8841e91aa465a5f95"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
