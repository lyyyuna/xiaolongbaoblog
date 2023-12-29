title: How to use Ginkgo label in your test
date: 2023-12-31 10:04:00
series: Ginkgo 使用笔记
summary: A Ginkgo label tutorial

---

Many testing frameworks provide the ability to group test cases, such as the `mark` feature in `pytest`, the `tag` feature in `Robot Framework`, and the `groups` feature in `TestNG`.

In `Ginkgo v1`, there is no similar feature available, we have to embed some keywords in the spec title and use the `--focus` option to filter specific test cases. This solution is not perfect. 