from django.urls import include, path
from rest_framework import routers

from sync.views import StatusViewSet

router = routers.DefaultRouter()
router.register('status', StatusViewSet, basename='status')

urlpatterns = [
    path('', include(router.urls)),
]
