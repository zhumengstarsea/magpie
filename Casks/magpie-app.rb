cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1734"
  sha256 arm:   "0da53eaeee1ae509d8ba4929893d2cb73214aae202c7d9c867c79495b1e84463",
         intel: "6c77270a479d1cedbd3e92f61f5b782b417398914b51c11f09ffc37310b75e91"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
