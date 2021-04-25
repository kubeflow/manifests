import io
import ruamel.yaml
from typing import List, Dict
from collections import UserDict


class YAMLEmitterNoVersionDirective(ruamel.yaml.emitter.Emitter):
    """YAML Emitter that doesn't emit the YAML version directive."""

    def write_version_directive(self, version_text):
        """Disable emitting version directive, i.e., %YAML 1.1."""
        pass

    def expect_document_start(self, first=False):
        """Do not print '---' at the beginning."""
        if not isinstance(self.event, ruamel.yaml.events.DocumentStartEvent):
            return super(YAMLEmitterNoVersionDirective, self).\
                expect_document_start(first=first)

        version = self.event.version
        self.event.version = None
        ret = super(YAMLEmitterNoVersionDirective, self).\
            expect_document_start(first=first)
        self.event.version = version
        return ret


class YAML(ruamel.yaml.YAML):
    """Wrapper of the ruamel.yaml.YAML class with our custom settings."""

    def __init__(self, *args, **kwargs):
        super(YAML, self).__init__(*args, **kwargs)
        # XXX: Explicitly set version for producing K8s compatible manifests.
        # https://yaml.readthedocs.io/en/latest/detail.html#document-version-support
        self.version = (1, 1)
        # XXX: Do not emit version directive since tools might fail to
        # parse manifests.
        self.Emitter = YAMLEmitterNoVersionDirective
        # Preserve original quotes
        self.preserve_quotes = True


yaml = YAML()


def lfilter(fn, iterable):
    return list(filter(fn, iterable))


class KubernetesObject(UserDict):
    """ A Kubernetes object.

    The class emulates a dict:
    https://docs.python.org/3/reference/datamodel.html?emulating-container-types#emulating-container-types
    """

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        if "apiVersion" not in self:
            raise ValueError("Invalid object, no apiVersion: %s" % repr(self))
        if "metadata" not in self:
            raise ValueError("Invalid object, no metadata: %s" % repr(self))
        # if "namespace" not in self["metadata"]:
        #     raise ValueError("Invalid object, no namespace: %s" % self)
        if "name" not in self["metadata"]:
            raise ValueError("Invalid object, no name: %s" % repr(self))

    def __hash__(self) -> int:
        hashstr = "%s|%s|%s|%s" % (self["apiVersion"], self["kind"],
                                   self["metadata"].get("namespace", ""),
                                   self["metadata"]["name"])
        return hash(hashstr)

    def __str__(self) -> str:
        msg = ("%s | %s | %s | %s " % (self["apiVersion"], self["kind"],
                                       self["metadata"].get("namespace", ""),
                                       self["metadata"]["name"]))
        return msg

    def __repr__(self) -> str:
        string_stream = io.StringIO()
        YAML().dump(self.data, stream=string_stream)
        output_str = string_stream.getvalue()
        string_stream.close()
        return output_str

    @staticmethod
    def from_meta(api_version, kind, name, namespace=None):
        obj = {"apiVersion": api_version, "kind": kind,
               "metadata": {"name": name}}
        if namespace:
            obj["metadata"]["namespace"] = namespace
        return KubernetesObject(obj)


class KubernetesObjectCollection(object):
    """ A collection of Kubernetes Objects.

    Provides utility functions to search and index Kubernetes Objects.
    """

    def __init__(self, objs: List[Dict]):
        # Validate objects
        self.objs = []
        for obj in objs:
            if obj:
                self.objs.append(KubernetesObject(obj))

    def list(self, api_version="*", kind="*", name="*", namespace="*",
             filter_fn=None):
        res = self.objs
        if api_version != "*":
            def has_api_version(obj): return obj["apiVersion"].casefold() == api_version.casefold()  # noqa:E501
            res = lfilter(has_api_version, res)
        if kind != "*":
            def has_kind(obj): return obj["kind"].casefold() == kind.casefold()
            res = lfilter(has_kind, res)
        if name != "*":
            def has_name(obj): return obj["metadata"]["name"] == name
            res = lfilter(has_name, res)
        if namespace != "*":
            def has_namespace(obj): return obj["metadata"].get("namespace", "") == namespace  # noqa:E501
            res = lfilter(has_namespace, res)
        if filter_fn:
            res = lfilter(filter_fn, res)
        return res

    def get(self, api_version, kind, name, namespace=None):
        for obj in self.objs:
            if api_version != obj["apiVersion"]:
                continue
            if kind != obj["kind"]:
                continue
            if name != obj["metadata"]["name"]:
                continue
            if namespace and namespace != obj["metadata"]["namespace"]:
                continue
            return obj
        raise RuntimeError("Object not found: %s | %s | %s | %s" %
                           (api_version, kind, name, namespace))

    def __iter__(self):
        return iter(self.objs)
