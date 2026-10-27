class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1661/magpie-cli-darwin-arm64"
      sha256 "45effe5526a73f8b282c6e394b6603328109689abcce833ed5aaec189f880212"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1661/magpie-cli-darwin-amd64"
      sha256 "ac64cad5f963cd7becd3c51697dc902dadc7c9df06d800ea1de8c597a842864d"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1661/magpie-cli-linux-arm64"
      sha256 "191ca0c831e1a67e3965cd5569b2ead15d647dc8d14745e96a2c6e5288de1e4a"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1661/magpie-cli-linux-amd64"
      sha256 "e8fcb52ca75572dd2b37bafcd6e400b83c4d1cde07797bdde9684455d40590e9"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
