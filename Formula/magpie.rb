class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1545/magpie-cli-darwin-arm64"
      sha256 "5d8978821ed65aaf5b575d5c71d439fa50c2ed6a7116ad7435357aa86e604168"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1545/magpie-cli-darwin-amd64"
      sha256 "491ff6af30b99a9525504f981b4bdbc609986fd16f78d2e151752d7d1e929be8"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1545/magpie-cli-linux-arm64"
      sha256 "501ac719e4541f60bdacefdeb672d2a477b387f86cbfffa0414f9870fbaa5755"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1545/magpie-cli-linux-amd64"
      sha256 "8516f61f5e3d7a2b3ef4275bd24141795f4241ec69f22f887d9c128dd95332cc"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
