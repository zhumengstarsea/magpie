cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1294"
  sha256 arm:   "4f11df9066c730a1846e9061ed648f19a6399a75493ae3d814503aa8d9934cb5",
         intel: "19c74e8104e2a2b595bc9646e7601622a78bb661fed1b0def4307baff7f1e9e9"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
