class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1689/magpie-cli-darwin-arm64"
      sha256 "5f89df3ff7d38f0409539a531c1c5c30493a1ac155e02e0a354d7340102f90ef"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1689/magpie-cli-darwin-amd64"
      sha256 "f30c6cb488db1d60a36f7c96c6f1e1954380fc4d33079ef02b1c1e89dafa3602"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1689/magpie-cli-linux-arm64"
      sha256 "16f2b3e5adca998f83973f2544738b98e62e96a5d181d4abc3e156864a0c7d6b"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1689/magpie-cli-linux-amd64"
      sha256 "abd1df8ba38669192d55a19c8953fd1502c60daadfe5628e9fef703a4976fdf1"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
