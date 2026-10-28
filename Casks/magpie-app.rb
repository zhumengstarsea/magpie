cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1841"
  sha256 arm:   "15559f7c59d4c67e9453877df3b3d11af440c4512accc90f9e40c720f463932f",
         intel: "5611fcd6057e65249c26d228a34c945771dec0c223a2bf81a107eda0ce1346f2"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
