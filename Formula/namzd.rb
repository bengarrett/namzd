class Namzd < Formula
  desc "Quickly find files by name or extension"
  homepage "https://github.com/bengarrett/namzd"
  url "https://github.com/bengarrett/namzd/archive/refs/tags/v1.3.1.tar.gz"
  sha256 "d0700979bcb6c0015c0d289791e8512db53be036439a541d46f2a4d61deea40a"
  version "1.3.1"
  license "GPL-3.0-only"

  @commit = "14e1b123985985ec1b29668e718381a32e1b74f5"
  @build_date = "2026-05-01T16:32:58+10:00"

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
