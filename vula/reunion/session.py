import os
from typing import Dict, List

from vula.peer import Peer

"""
This module implements the REUNION cryptographic protocol as described in
Algorithm 1 of the REUNION paper.
"""

# Curve25519 E/F p with Diﬀie-Hellman function DH(private, public), base point
# P ∈ E(F p ), 32-byte EpochID, 32-byte SharedRandom, aead-enc(key, plaintext,
# ad) function, aead-dec(key, ciphertext, ad) function, password hashing
# function argon2id(), key derivation function HKDF(), 32-byte block cipher
# rijndael-enc(key, plaintext)/rijndael-dec(key, ciphertext), random number
# generator function RNG, Elligator encode/decode functions.

from reunion.primitives import (
    x25519,
    PrivateKey,
    PublicKey,
    csidh,
    aead_encrypt,
    aead_decrypt,
    argon2i,
    Hkdf,
    prp_encrypt,
    prp_decrypt,
    elligator,
    unelligator,
    generate_elligatorable_sk,
    Hash,
    blake2b,
)

DEFAULT_SALT = bytes("\0".encode() * 32)


class T1(bytes):

    """
    This class provides attribute access to the four parts of a T1 message, and
    its id.

    The T1 message consists of four sections; the α, β, γ, and δ:
    – α is an Elligator-encoded Curve25519 public key, encrypted by a PRP using
    a symmetric key derived as described in Step 8. The encryption and
    decryption of α is unauthenticated, i.e. does not provide a validity check.
    The elligator encoding ensures that plaintext is indistinguishable from
    random bytes and every sequence of random bytes maps to a valid public key.
    – β is a CSIDH public key
    – γ is the MAC of an AEAD encryption of an empty string using a random
    key which is revealed by the T2.
    – δ is an AEAD ciphertext containing the message payload which is only de-
    cryptable by a valid T3.
    """

    LEN_ALPHA = 32
    LEN_BETA = 64
    LEN_GAMMA = 16

    @property
    def alpha(self):
        return self[: self.LEN_ALPHA]

    @property
    def beta(self):
        return self[self.LEN_ALPHA : self.LEN_ALPHA + self.LEN_BETA]

    @property
    def gamma(self):
        return self[
            self.LEN_ALPHA
            + self.LEN_BETA : self.LEN_ALPHA
            + self.LEN_BETA
            + self.LEN_GAMMA
        ]

    @property
    def delta(self):
        return self[self.LEN_ALPHA + self.LEN_BETA + self.LEN_GAMMA :]

    @property
    def id(self):
        return Hash(self)

    def __repr__(self):
        return "<%s:%s>" % (
            type(self).__name__,
            self.id[:6].hex(),
        )


class ReunionSession(object):

    """
    Phases 0 and 1 are implemented in this object directly, while 2, 3, and 4
    are implemented in the Peer object which ReunionSession instantiates for
    each new t1.
    """

    @classmethod
    def keygen(cls):
        """
        All of the RNG access is consolodated here in a class method, so that
        everything else in the ReunionSession and Peer classes (except "create"
        which calls this) can be sans IO.
        """
        return dict(
            dh_sk=generate_elligatorable_sk(),
            csidh_sk=csidh.secret_key(),
            gamma_seed=Hash(os.urandom(32)),
            delta_seed=Hash(os.urandom(32)),
            dummy_seed=Hash(os.urandom(32)),
            tweak=Hash(os.urandom(32))[0],
        )

    @classmethod
    def create(cls, passphrase: bytes, payload: bytes, salt=DEFAULT_SALT):
        """
        This is the typical way to instantiate a ReunionSession object.
        """
        return cls(
            payload=payload,
            salt=salt,
            basekey=argon2i(passphrase, salt),
            **cls.keygen(),
        )

    def __init__(
        self,
        salt: bytes,
        basekey: bytes,
        payload: bytes,
        dh_sk: bytes,
        csidh_sk: bytes,
        gamma_seed: bytes,
        delta_seed: bytes,
        dummy_seed: bytes,
        tweak: int,
    ):
        # dict of Peer objects, keyed by their t1 id
        self.peers: Dict = {}

        # list of payloads decrypted
        self.results: List = []

        # Step2a: esk Aα ∈ Z, public key epk Aα = esk Aα · P ∈ E(Fp).
        self.dh_sk: PrivateKey = PrivateKey(dh_sk)
        dh_pk: PublicKey = self.dh_sk.public_key

        # Step2b: esk Aβ ∈ Z, public key epk Aβ = esk Aβ · P ∈ E(Fp)
        self.csidh_sk: bytes = csidh_sk
        csidh_pk: bytes = csidh.public_key(csidh_sk)

        # Step 3: salt ← SharedRandom∥EpochID.
        # Setting the salt, or context, is the responsibility of the
        # application using this library. DEFAULT_SALT is 32 null bytes.
        self.salt = salt

        # Step 4: pdk ← HKDF(salt, argon2id(salt, Q)).
        kdf = Hkdf(salt=salt, input_key_material=basekey, hash=blake2b)
        self.pdk = kdf.expand(b"", 32)

        # Step 5a: sk Aγ ← H(pdk, RNG(32), msg A )
        self.sk_gamma = Hash(self.pdk + gamma_seed + payload)

        # Step 5b: sk Aδ ← H(pdk, RNG(32), msg A )
        self.sk_delta = Hash(self.pdk + delta_seed + payload)

        # t1 beta is the unencrypted csidh pk
        beta = csidh_pk

        # Step 6: T1Aγ ← aead-enc(sk Aγ ,“”, RS)
        gamma = aead_encrypt(self.sk_gamma, b"", salt)

        # Step 7: T1Aδ ← aead-enc(sk Aδ , msg a , RS)
        delta = aead_encrypt(self.sk_delta, payload, salt)

        # Step 8: pdkA ← H(pdk, epkAβ , T1Aγ , T1Bδ )
        self.alpha_key = Hash(self.pdk + csidh_pk + gamma + delta)

        # Step 9: T1Aα ← rijndael-enc(pdk A , elligator(epk Aα))
        alpha = prp_encrypt(self.alpha_key, elligator(dh_pk.encode(), tweak))

        # Step 10: T1A ← T1 Aα ∥ epkAβ ∥ T1 Aγ ∥ T1 Aδ
        self.t1 = T1(alpha + beta + gamma + delta)

        # deviation from the paper - we generate dummys derministically using this Hkdf.
        self.dummy_hkdf = Hkdf(salt=salt, input_key_material=dummy_seed, hash=blake2b)

    # steps 11, 12, and 13 happen in the application using this library

    def process_t1(self, t1: bytes):
        # Step 14: for each new T1Bi do ▷ Phase 2: Process T1; transmit T2
        # implementation of steps 14 to 22 is in Peer.__init__
        if t1 == self.t1:
            assert (
                False
            ), "we don't actually expect this to happen right now but later might return None here"
            return None

        peer = self.peers.get(t1.id)
        if peer is None:
            peer = SessionPeer(t1, self)
            self.peers[t1.id] = peer

        return peer.t2_tx

    # steps 23 and 24, transmit t1, is in the application

    def process_t2(self, t1_id: bytes, t2: bytes):
        # Step 25: for each new T2Bi do ▷ Phase 3: Process T2, transmit T3
        # implementation of steps 25 to 33 is in Peer.process_t2
        if t1_id in self.peers:
            return self.peers[t1_id].process_t2(t2)
        else:
            return self.dummy_hkdf.expand(t1_id + t2, 32), True

    def process_t3(self, t1_id: bytes, t3: bytes):
        # Step 36: for each new T3Bi do ▷ Phase 4: Process T3; decrypt δ
        if t1_id in self.peers:
            return self.peers[t1_id].process_t3(t3)


class SessionPeer(object):
    def __init__(peer, t1: T1, session: ReunionSession):
        """
        This method implements the inside of the for loop in Phase 2 of
        Algorithm 1.
        """

        # Step 14: for each new T1Bi do ▷ Phase 2: Process T1; transmit T2

        peer.t1 = t1
        peer.session = session

        # Step 15: pdkBi ← H(pdk, T1Biβ, T1Biγ, T1Biδ)
        peer.alpha_key = Hash(session.pdk + t1.beta + t1.gamma + t1.delta)

        # Step 16: epkBiα ← unelligator(rijndael-dec(pdkBi , T1Biα )).
        dh_pk: PublicKey = PublicKey(unelligator(prp_decrypt(peer.alpha_key, t1.alpha)))

        # Step 17: epkBiβ ← T1Biβ
        csidh_pk = t1.beta

        # Step 18: dh1ssi ← H(DH(eskAα , epkBiα))
        dh_ss = x25519(session.dh_sk, dh_pk)

        # Step 19: dh2ssi ← H(DH(eskAβ , epkBiβ)).
        csidh_ss = csidh.dh(session.csidh_sk, csidh_pk)

        # Step 20: T2kitx ← H(pdkA, pdkBi, dh1ssi, dh2ssi)
        peer.t2key_tx = Hash(session.alpha_key + peer.alpha_key + dh_ss + csidh_ss)

        # Step 21: T2kirx ← H(pdkBi, pdkA, dh1ssi, dh2ssi)
        peer.t2key_rx = Hash(peer.alpha_key + session.alpha_key + dh_ss + csidh_ss)

        # Step 22: T2Ai ← rijndael-enc(T2kitx, skAγ)
        peer.t2_tx = prp_encrypt(peer.t2key_tx, session.sk_gamma)
        peer.t2_rx = None
        peer.payload = None

        # Step 23 and 24, transmit, is implemented in an application using this library

    def process_t2(self, peer, t2):
        """
        This method implements the inside of the for loop in Phase 3 of
        Algorithm 1. It returns a 2-tuple of (t3, is_dummy).
        """

        # Step 25: for each new T2Bi do ▷ Phase 3: Process T2, transmit T3

        # Step 26: skBiγ ← rijndael-dec(T2kirx, T2Bi)
        sk_gamma = prp_decrypt(peer.t2key_rx, t2)

        # Step 27: if “” = aead-dec(sk B i γ , T 1 B i γ , RS) then
        aead_res = aead_decrypt(sk_gamma, peer.t1.gamma, peer.session.salt)

        if aead_res is not None:
            assert aead_res == b"", aead_res

            peer.t2_rx = t2

            # Step 28: T3kitx ← H(T2kitx, T2Ai , T2Bi).
            t3key_tx = Hash(peer.t2key_tx + peer.t2_tx + t2)

            # Step 29: T3kirx ← H(T2kirx, T2Bi, T2Ai)
            self.t3_key_rx = Hash(peer.t2key_rx + peer.t2_rx + peer.t2_tx)

            # Step 30: T3Ai ← rijndael-enc(T3kitx, skAδ)
            return prp_encrypt(t3key_tx, peer.session.sk_delta), False

        # Step 31: else
        else:
            # Step 32: T3Ai ← H(RNG(32))
            return peer.session.dummy_hkdf.expand(peer.t1.id + t2, 32), True

    def process_t3(self, peer, t3: bytes):

        """
        This method implements the inside of the for loop in Phase 4 of
        Algorithm 1.
        """

        # Step 36: for each new T3Bi do ▷ Phase 4: Process T3; decrypt δ

        if peer.t2_rx is None:
            # we didn't receive a valid t2 from this peer, so, we can't decrypt
            # their t3.
            # At the cost of a more complicated API, we could implement support
            # for out-of-order delivery by storing t3s here and attempting to
            # process them again when we receive a t2 from the corresponding
            # t1. As it is now, if we receive their t3 before their t2 we will
            # never decrypt the t1 delta payload.
            return None

        # Step 37: skBiδ ← rijndael-dec(T3kirx, T3Bi).
        sk_delta = prp_decrypt(self.t3_key_rx, t3)

        # Step 38: if msgBi ← aead-dec(skBiδ, T1Biδ, RS) then
        peer.payload = aead_decrypt(sk_delta, peer.t1.delta, peer.session.salt)
        if peer.payload is not None:
            # Step 39: add to results
            peer.session.results.append(peer.payload)
        return peer.payload
