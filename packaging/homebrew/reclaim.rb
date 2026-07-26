# Homebrew formula TEMPLATE for reclaim — do not edit the placeholders.
#
# The `homebrew` job in .github/workflows/release.yml fills in VERSION and
# the SHA256 placeholders from the release artifacts, then pushes the result
# to ImadRashid/homebrew-tap as Formula/reclaim.rb (requires the
# TAP_GITHUB_TOKEN repo secret). No manual steps.
class Reclaim < Formula
  desc "Developer-aware Mac cleaner CLI"
  homepage "https://github.com/ImadRashid/reclaim"
  version "VERSION"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/ImadRashid/reclaim/releases/download/v#{version}/reclaim_v#{version}_darwin_arm64.tar.gz"
      sha256 "REPLACE_WITH_DARWIN_ARM64_SHA256"
    end
    on_intel do
      url "https://github.com/ImadRashid/reclaim/releases/download/v#{version}/reclaim_v#{version}_darwin_amd64.tar.gz"
      sha256 "REPLACE_WITH_DARWIN_AMD64_SHA256"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/ImadRashid/reclaim/releases/download/v#{version}/reclaim_v#{version}_linux_arm64.tar.gz"
      sha256 "REPLACE_WITH_LINUX_ARM64_SHA256"
    end
    on_intel do
      url "https://github.com/ImadRashid/reclaim/releases/download/v#{version}/reclaim_v#{version}_linux_amd64.tar.gz"
      sha256 "REPLACE_WITH_LINUX_AMD64_SHA256"
    end
  end

  def install
    bin.install "reclaim"
    doc.install "README.md"
  end

  test do
    assert_match "reclaim #{version}", shell_output("#{bin}/reclaim --version")
  end
end
