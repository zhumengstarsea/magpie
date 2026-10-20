cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1302"
  sha256 arm:   "df5dd4f23fd4930069dd7fe250cdef556bc5c0abce9384923d77c2826f9835d9",
         intel: "33e337837d4bf8205745a33888d55e62818f0c5ca10b8d544933bbb8e9414993"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
