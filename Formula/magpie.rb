class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1488/magpie-cli-darwin-arm64"
      sha256 "a6f7c117efb25f2427d875d23b61f469bf4e46d082e8ee1dec6574995f7be1b3"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1488/magpie-cli-darwin-amd64"
      sha256 "c7b242ff9e47760c779f5a21eb0730ac17f908215a0d2cba865565fdfa67ed05"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1488/magpie-cli-linux-arm64"
      sha256 "1cc8afba31ac962f544f9fc4ac5cdacdab50f9423b75001ca74a3c7a5106d513"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1488/magpie-cli-linux-amd64"
      sha256 "d9c73003cc2859217d3d1568f41ad896961a1f48e7f78dc7d0746ffd037d5d99"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
