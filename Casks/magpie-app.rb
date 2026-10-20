cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1296"
  sha256 arm:   "def8b7084463195f1756c18261f1194b2cb2ecdc4ee51f631e3181dff4cd94b6",
         intel: "caf095416878424da6c7c3c4d2f373e73ee614b34502c55ee85e11df8cae83e1"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
