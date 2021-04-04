import React from 'react';
import { Container, Row } from 'react-bootstrap';
import {
  faChalkboardTeacher,
  faQuestionCircle,
  faUsers,
} from '@fortawesome/free-solid-svg-icons';
import ControlButton from './ControlButton';
import logo from '../../../assets/homeHeader.svg';

const Header = () => (
  <div className="Header">
    <div className="Logo">
      <img alt="logo" src={logo} />
    </div>
    <div className="ControlBar">
      <Container>
        <Row className="Controls">
          <ControlButton
            name="team"
            action={() => console.log('team button clicked')}
            icon={faUsers}
          />

          <ControlButton
            name="boards"
            action={() => console.log('boards button clicked')}
            icon={faChalkboardTeacher}
          />

          <ControlButton
            name="help"
            action={() => console.log('boards button clicked')}
            icon={faQuestionCircle}
          />
        </Row>
      </Container>
    </div>
  </div>
);

export default Header;
