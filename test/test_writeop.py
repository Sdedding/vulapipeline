import unittest

from vula.organize import OrganizeState, SystemState
from vula.engine import WriteOp


class TestWriteOps(unittest.TestCase):
    def setUp(self):
        self.state = OrganizeState()
        # initialize system state and a local domain
        self.state.event_NEW_SYSTEM_STATE(
            SystemState(current_subnets={"10.0.0.0/24": ["10.0.0.9"]})
        )
        self.state.event_USER_EDIT('SET', 'prefs.local_domains', ['local'])

    def test_set_writeop(self):
        res = self.state.event_USER_EDIT('SET', 'prefs.record_events', True)
        self.assertIsInstance(res.writes[0], WriteOp)
        self.assertEqual(
            (res.writes[0].kind, res.writes[0].path, res.writes[0].value),
            ('SET', ('prefs', 'record_events'), True),
        )
        self.assertTrue(self.state.prefs.record_events)

    def test_add_writeop(self):
        res = self.state.event_USER_EDIT(
            'ADD', ['prefs', 'local_domains'], 'example.net'
        )
        self.assertIsInstance(res.writes[0], WriteOp)
        self.assertEqual(res.writes[0].kind, 'ADD')
        self.assertIn('example.net', self.state.prefs.local_domains)

    def test_remove_writeop(self):
        res = self.state.event_USER_EDIT(
            'REMOVE', ['prefs', 'local_domains'], 'local'
        )
        self.assertIsInstance(res.writes[0], WriteOp)
        self.assertEqual(res.writes[0].kind, 'REMOVE')
        self.assertNotIn('local', self.state.prefs.local_domains)


if __name__ == '__main__':
    unittest.main()
