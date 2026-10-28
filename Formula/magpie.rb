class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1841/magpie-cli-darwin-arm64"
      sha256 "87a537c936676fc81ffad3bfcc6547830d88f8c334a5383b687ef7b2e38f7751"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1841/magpie-cli-darwin-amd64"
      sha256 "dffbbd0d6e2b05b82b65e099182ba297be78763cfd1db44babac218140ef566f"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1841/magpie-cli-linux-arm64"
      sha256 "f6fc98e3d5621c964f8101a5e785f8b550d0db14b98272eb124e9a5859e0c380"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1841/magpie-cli-linux-amd64"
      sha256 "426df8a57eeb86a7e9370ecf11d125fcf0bc5833dfb1bf726d2d3f6aaea1527e"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
