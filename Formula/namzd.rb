class Namzd < Formula
  desc "Quickly find files by name or extension"
  homepage "https://github.com/bengarrett/namzd"
  url "https://github.com/bengarrett/namzd/archive/refs/tags/v1.3.0.tar.gz"
  sha256 "17446067087fafdbfdcf207bd3d1b52d1f872e736f6a7c05b2666018b08bb582"
  version "1.3.0"
  license "GPL-3.0-only"

  @commit = "da221254381fc2079b739d2c5611d3c605d86eec"
  @build_date = "2026-02-06T12:11:20+11:00"

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
