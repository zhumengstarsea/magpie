cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1539"
  sha256 arm:   "ef120d29afbf2242bee98c6ba5b7d99b534776cbb3d7a5080be0eef02779f3bc",
         intel: "3edc64a1f8f1e24442e77943ed0a4264b54772d1651050b02092c7b9e89b3bdc"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
