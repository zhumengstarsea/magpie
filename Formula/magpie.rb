class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1957/magpie-cli-darwin-arm64"
      sha256 "3d4d486e33a3d60d4ae2c07fa745484af1bcc5807ad99d03eda64cd701649e2a"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1957/magpie-cli-darwin-amd64"
      sha256 "f26af033ef486eb446c56a7c4afde2eb58f44e6c7fde006e25e6f0a86e19edd7"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1957/magpie-cli-linux-arm64"
      sha256 "358f713e0ba67aa54a697ff915514025a918cce77b84181c640a03375719018c"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1957/magpie-cli-linux-amd64"
      sha256 "2c21f5018e36c35c2cdc21551b0b231ecf31fee459c957a2c807ad82f4a57273"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
