class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1306/magpie-cli-darwin-arm64"
      sha256 "8a53ccd340a1586e75260951aa952d5ee6325df97545c3a80d1a85ec2f0ed449"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1306/magpie-cli-darwin-amd64"
      sha256 "539dac4efd0d5b2f71e8dcdd0bce68faeacfee0dca4d912aea68b8c6729a1926"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1306/magpie-cli-linux-arm64"
      sha256 "b23cf10aff19fba7a2828733c2316bc0d0670a251f0d1df5f474ef1515c7ec61"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1306/magpie-cli-linux-amd64"
      sha256 "892c1eff7f4b29205c64b0459eb13ffbb3c988ea537f480b2064a8bf601723c9"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
