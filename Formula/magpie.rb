class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1616/magpie-cli-darwin-arm64"
      sha256 "9393dab2a1c0064b306c773b6df6f94bc3d85dd23656e5e24f56fcf4c2a5a06d"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1616/magpie-cli-darwin-amd64"
      sha256 "3e98cd4fcf9f046e49f593c5c03ef155d3da86e0766355660cea192a761e0daf"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1616/magpie-cli-linux-arm64"
      sha256 "f6537bd9f5ec3b572a7dd4ecfd2310c0943e70e05beee1185e3f66ad01c1b83f"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1616/magpie-cli-linux-amd64"
      sha256 "8dbeda0b60491ac512b01c2f00522ad7dc4d3ea0a588b1ad7510574abe1f2ca5"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
