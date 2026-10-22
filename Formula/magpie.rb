class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1322/magpie-cli-darwin-arm64"
      sha256 "1075f3a7ce821c30eb44b4951743e551ffcad3eb6b7fdeab34a0039a7ab32ca8"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1322/magpie-cli-darwin-amd64"
      sha256 "5b18d33575274c0a3f45df387831479ec40ae858f72381c39ffd535aaebdc6d6"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1322/magpie-cli-linux-arm64"
      sha256 "b0205d6e43096fe36ed2cdd202c707b0f85b7fe6025009386386f3c382e5059b"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1322/magpie-cli-linux-amd64"
      sha256 "117bbce0beb0543b5aa2607d53ccd839c74755e199b89ceeb6282edb729a311a"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
