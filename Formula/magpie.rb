class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1311/magpie-cli-darwin-arm64"
      sha256 "c99dfde6097ff6c1724a1d959a694ce80997ca05e71b2b340e51d736bad88f71"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1311/magpie-cli-darwin-amd64"
      sha256 "fcbc88e6fa2d49b4faef9a5f7f36c3a6d6e0a825f1f3c3f2ac39305a1585ecb3"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1311/magpie-cli-linux-arm64"
      sha256 "f860c33753fc6be6853fe0e5464bac49c004bb59ebee7b87ec1e5d93e5cf10d9"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1311/magpie-cli-linux-amd64"
      sha256 "749e92a21982f7cc0adab0e519b7dc64b666a0d5d0de71f16b4635986b62ae13"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
