class Namzd < Formula
  desc "Quickly find files by name or extension"
  homepage "https://github.com/bengarrett/namzd"
  url "https://github.com/bengarrett/namzd/archive/refs/tags/v1.2.4.tar.gz"
  sha256 "29c25beebb1b69037c6877332181fda8e1d0c60a0f4390476f155bde7f652bf6"
  version "1.2.4"
  license "GPL-3.0-only"

  @commit = "7c36e0920da66d7df404926c90c997708f88fedb"
  @build_date = "2026-02-05T08:06:19Z"

  livecheck do
    url :stable
    strategy :github_latest
  end

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w -X main.version=#{version} -X main.commit=#{self.class.instance_variable_get('@commit')} -X main.date=#{self.class.instance_variable_get('@build_date')}")
  end

  test do
    assert_match "namzd", shell_output("#{bin}/namzd --version")
  end
end
