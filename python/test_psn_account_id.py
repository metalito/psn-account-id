"""Tests for psn_account_id.py. Run:  python3 -m unittest -v  (from this folder)."""
import io
import json
import unittest
from unittest import mock

import psn_account_id as p

USER_ID = "1234567890123456789"
B64 = "FYHpffQQIhE="


class ConvertTests(unittest.TestCase):
    def test_number_to_base64(self):
        self.assertEqual(p.convert(USER_ID)["account_id_base64"], B64)

    def test_base64_to_number(self):
        self.assertEqual(p.convert(B64)["account_id_number"], USER_ID)

    def test_rejects_wrong_length(self):
        with self.assertRaises(p.PsnError):
            p.convert("AAAA")

    def test_rejects_too_large(self):
        with self.assertRaises(p.PsnError):
            p.convert(str(2 ** 64))


class ExtractCodeTests(unittest.TestCase):
    def test_full_url(self):
        self.assertEqual(p.extract_code(p.REDIRECT + "?code=abc123&cid=x"), "abc123")

    def test_bare_code_and_quotes(self):
        self.assertEqual(p.extract_code('  "abc123" '), "abc123")

    def test_rejects_other_site(self):
        with self.assertRaises(p.PsnError):
            p.extract_code("https://evil.example/remoteplay/redirect?code=abc")

    def test_rejects_missing_code(self):
        with self.assertRaises(p.PsnError):
            p.extract_code(p.REDIRECT + "?foo=1")


class LoginUrlTests(unittest.TestCase):
    def test_points_at_sony(self):
        url = p.login_url()
        self.assertTrue(url.startswith(p.AUTH_HOST + "/2.0/oauth/authorize?"))
        self.assertIn(p.CLIENT_ID, url)


class FlowTests(unittest.TestCase):
    def test_fetch_account_only_calls_sony(self):
        calls = []

        def fake(url, form=None):
            calls.append(url)
            if form is not None:
                self.assertEqual(form["code"], "abc")
                return {"access_token": "AT"}
            return {"user_id": USER_ID, "online_id": "Player1"}

        with mock.patch.object(p, "sony_request", side_effect=fake):
            result = p.fetch_account("abc")
        self.assertEqual(result["account_id_base64"], B64)
        self.assertEqual(result["online_id"], "Player1")
        self.assertTrue(all(u.startswith(p.AUTH_HOST + "/") for u in calls))

    def test_main_json(self):
        with mock.patch.object(p, "fetch_account", return_value={"account_id_base64": B64}), \
             mock.patch("builtins.input", return_value="abc"), \
             mock.patch("sys.stderr", new_callable=io.StringIO), \
             mock.patch("sys.stdout", new_callable=io.StringIO) as out:
            self.assertEqual(p.main(["--json", "--no-browser"]), 0)
        self.assertEqual(json.loads(out.getvalue())["account_id_base64"], B64)

    def test_main_convert(self):
        with mock.patch("sys.stdout", new_callable=io.StringIO) as out:
            self.assertEqual(p.main(["--convert", USER_ID]), 0)
        self.assertIn(B64, out.getvalue())


if __name__ == "__main__":
    unittest.main()
