class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1307/magpie-cli-darwin-arm64"
      sha256 "39f95ee88b13db4be95354ba7e93c477d02117bad60a4b84c69b16c35d3ca26e"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1307/magpie-cli-darwin-amd64"
      sha256 "f50cb50463967ed59b212c5cb13dc021ae4b2b0b8bab85b395dea88b783c17b0"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1307/magpie-cli-linux-arm64"
      sha256 "7fc9c650e534031888d58956c7b90d8d99e17a12f93793fe775413cdb092d006"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1307/magpie-cli-linux-amd64"
      sha256 "af49df0289e66471cc25705614a1999d08f0c7716976cbf6794ea7c1d0011160"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
