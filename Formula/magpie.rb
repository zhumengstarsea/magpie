class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1318/magpie-cli-darwin-arm64"
      sha256 "61cb8ee703bc576c3d0043de32a1a539767680f96a54f7b43dd8341d80fc1938"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1318/magpie-cli-darwin-amd64"
      sha256 "d6fbb0647ead8e7578cf1a376abc0e5f6d835404b084fdcc0912c81d7ad5c5aa"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1318/magpie-cli-linux-arm64"
      sha256 "6755596086145ea74bd66311b2a12b76aa1a6e4305918e866a4e2fd2e4d74f23"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1318/magpie-cli-linux-amd64"
      sha256 "415016e086f562603205820a25ac9412e42c876977d351e47aa6f6c2cc226be9"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
