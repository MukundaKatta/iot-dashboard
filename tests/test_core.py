"""Tests for IotDashboard."""
from src.core import IotDashboard
def test_init(): assert IotDashboard().get_stats()["ops"] == 0
def test_op(): c = IotDashboard(); c.detect(x=1); assert c.get_stats()["ops"] == 1
def test_multi(): c = IotDashboard(); [c.detect() for _ in range(5)]; assert c.get_stats()["ops"] == 5
def test_reset(): c = IotDashboard(); c.detect(); c.reset(); assert c.get_stats()["ops"] == 0
def test_service_name(): c = IotDashboard(); r = c.detect(); assert r["service"] == "iot-dashboard"
