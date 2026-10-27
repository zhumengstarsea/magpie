class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1741/magpie-cli-darwin-arm64"
      sha256 "4ea051ce60db2bdb9c41371dc3c7b56ed2f06689d21aecc95d0376e6df283141"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1741/magpie-cli-darwin-amd64"
      sha256 "04a7093b6af88c1ac85acaffbbcb575666daf112be28734928df79055ab26ffa"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1741/magpie-cli-linux-arm64"
      sha256 "6cc75ee0385717d114a3db3141472079e6ff4e5e054d520a581a8443926fc16a"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1741/magpie-cli-linux-amd64"
      sha256 "eae487fc3ad2272b7d545254857db3243624dd6e472082f074ec22f9ded333bd"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
