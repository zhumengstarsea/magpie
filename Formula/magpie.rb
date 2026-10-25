class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1559/magpie-cli-darwin-arm64"
      sha256 "5c5f05229b98a79b1ed55de0d547dd39f7451c5054574056c2d5e2b7bc140b05"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1559/magpie-cli-darwin-amd64"
      sha256 "f55aaa94acf6a5c89c438090623cafe7f05f23d52ce735d13175c17ceb816810"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1559/magpie-cli-linux-arm64"
      sha256 "2c6f427e10e63649485999ae35f90ac792412aa56ed65b56d11046c5fdaa1553"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1559/magpie-cli-linux-amd64"
      sha256 "8355550be2dc6b9be0baca6dd052726753b0d855d12a3375afbd383829519261"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
