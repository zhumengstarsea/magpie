cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1485"
  sha256 arm:   "e711b568fa77849a434c0ad675606121e3b230ade53f8aea3f3eaf9bda7c2a9c",
         intel: "da53b29184afc9d4d0faafcb654c5d2c69523a0227857878e207bc639e8a4380"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
